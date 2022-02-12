package stats

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"runtime"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/xpetit/fizzbuzz/v5"
)

type Service interface {
	Increment(cfg *fizzbuzz.Config) error
	MostFrequent() (count int, cfg *fizzbuzz.Config, err error)
}

var (
	_ Service = (*Memory)(nil)
	_ Service = (*DB)(nil)
)

// Memory holds a protected (thread safe) hit count.
type Memory struct {
	mu sync.RWMutex
	m  map[fizzbuzz.Config]int
}

func (s *Memory) Increment(cfg *fizzbuzz.Config) error {
	if cfg == nil {
		cfg = &fizzbuzz.Config{}
	}
	s.mu.Lock()
	if s.m == nil {
		s.m = map[fizzbuzz.Config]int{}
	}
	s.m[*cfg]++
	s.mu.Unlock()
	return nil
}

func (s *Memory) MostFrequent() (count int, cfg *fizzbuzz.Config, err error) {
	s.mu.RLock()
	for config, c := range s.m {
		if c > count {
			count = c
			config := config
			cfg = &config
		} else if c == count && config.LessThan(cfg) {
			// Same hit count, the configs are differentiated because the "iteration order over maps is not specified" (Go spec)
			config := config
			cfg = &config
		}
	}
	s.mu.RUnlock()
	return
}

// DB holds a persistent and protected (thread safe) hit count.
// It must be closed when it is no longer needed.
type DB struct {
	ctx          context.Context
	db           *sql.DB
	insert       *sql.Stmt
	increment    *sql.Stmt
	mostFrequent *sql.Stmt
}

func Open(ctx context.Context, dataSourceName string) (*DB, error) {
	// Open database
	db, err := sql.Open("sqlite3", dataSourceName+"?"+url.Values{
		"_busy_timeout":        {"5000"},
		"_foreign_keys":        {"true"},
		"_journal_mode":        {"wal"},
		"_synchronous":         {"normal"},
		"_case_sensitive_like": {"true"},
	}.Encode())
	if err != nil {
		return nil, err
	}

	// Adjust database/sql settings to SQLite
	// This avoids several problems: ever-growing WAL file, file handle exhaustion, etc...
	if strings.Contains(dataSourceName, "memory") {
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(runtime.NumCPU())
		db.SetMaxIdleConns(runtime.NumCPU())
	}

	// Improve SQLite performance and reliability
	if _, err := db.ExecContext(ctx, `
		pragma wal_autocheckpoint = 0;
		pragma temp_store         = memory;
	`); err != nil {
		return nil, err
	}

	// Initialize database
	if _, err := db.ExecContext(ctx, `
		create table if not exists "stat" (
			"limit" integer not null,
			"int1"  integer not null,
			"int2"  integer not null,
			"str1"  text    not null,
			"str2"  text    not null,
			"count" integer not null,
			primary key (
				"limit",
				"int1",
				"int2",
				"str1",
				"str2"
			)
		) strict, without rowid;
		create index if not exists "idx_stat_count" on "stat" ("count");
	`); err != nil {
		return nil, err
	}

	// Prepare statements
	insert, err := db.PrepareContext(ctx, `
		insert into "stat" (
			"limit",
			"int1",
			"int2",
			"str1",
			"str2",
			"count"
		) values (
			?,
			?,
			?,
			?,
			?,
			1
		);
	`)
	if err != nil {
		return nil, err
	}
	increment, err := db.PrepareContext(ctx, `
		update
			"stat"
		set
			"count" = "count" + 1
		where
			(
				"limit",
				"int1",
				"int2",
				"str1",
				"str2"
			) = (
				?,
				?,
				?,
				?,
				?
			);
	`)
	if err != nil {
		return nil, err
	}
	mostFrequent, err := db.PrepareContext(ctx, `
		select
			"limit",
			"int1",
			"int2",
			"str1",
			"str2",
			"count"
		from
			"stat"
		where
			"count" = (select max("count") from "stat")
		order by
			"limit",
			"int1",
			"int2",
			"str1",
			"str2"
		limit 1;
	`)
	if err != nil {
		return nil, err
	}

	return &DB{
		ctx:          ctx,
		db:           db,
		insert:       insert,
		increment:    increment,
		mostFrequent: mostFrequent,
	}, nil
}

var (
	i  int
	m  sync.Mutex
	wg sync.WaitGroup
)

func (s *DB) Increment(cfg *fizzbuzz.Config) error {
	m.Lock()
	i++
	if i == 100_000 {
		i = 0
		wg.Wait()
		var failed bool
		if err := s.db.QueryRow(`pragma wal_checkpoint(restart)`).Scan(&failed, new(int), new(int)); err != nil {
			m.Unlock()
			return err
		}
		if failed {
			log.Println("pragma wal_checkpoint(restart) failed")
		}
	}
	m.Unlock()
	wg.Add(1)
	defer wg.Done()
	if cfg == nil {
		cfg = &fizzbuzz.Config{}
	}
	tx, err := s.db.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	result, err := tx.StmtContext(s.ctx, s.increment).ExecContext(s.ctx,
		cfg.Limit,
		cfg.Int1,
		cfg.Int2,
		cfg.Str1,
		cfg.Str2,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	nb, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if nb == 0 {
		if _, err := tx.StmtContext(s.ctx, s.insert).ExecContext(s.ctx,
			cfg.Limit,
			cfg.Int1,
			cfg.Int2,
			cfg.Str1,
			cfg.Str2,
		); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *DB) MostFrequent() (count int, cfg *fizzbuzz.Config, err error) {
	cfg = &fizzbuzz.Config{}
	err = s.mostFrequent.QueryRowContext(s.ctx).Scan(
		&cfg.Limit,
		&cfg.Int1,
		&cfg.Int2,
		&cfg.Str1,
		&cfg.Str2,
		&count,
	)
	if err == sql.ErrNoRows {
		return 0, nil, nil
	}
	return
}

func (s *DB) Close() error {
	if err := s.insert.Close(); err != nil {
		return err
	}
	if err := s.increment.Close(); err != nil {
		return err
	}
	if err := s.mostFrequent.Close(); err != nil {
		return err
	}
	return s.db.Close()
}
