package stats

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"runtime"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/xpetit/fizzbuzz/v5"
)

type db struct {
	ctx          context.Context
	db           *sql.DB
	increment    *sql.Stmt
	mostFrequent *sql.Stmt
}

// OpenDB opens a database holding a persistent and protected (thread safe) hit count.
// It must be closed when it is no longer needed.
func OpenDB(ctx context.Context, dataSourceName string) (*db, error) {
	sqlite, err := sql.Open("sqlite3", dataSourceName+"?"+url.Values{
		"_busy_timeout":        {"5000"},
		"_foreign_keys":        {"true"},
		"_journal_mode":        {"wal"},
		"_synchronous":         {"normal"},
		"_case_sensitive_like": {"true"},
		"_cache_size":          {"-512000"},
	}.Encode())
	if err != nil {
		return nil, err
	}

	// Adjust database/sql settings to SQLite to avoid ever-growing WAL file
	if strings.Contains(dataSourceName, "memory") {
		sqlite.SetMaxOpenConns(1)
	} else {
		sqlite.SetMaxOpenConns(runtime.NumCPU())
		sqlite.SetMaxIdleConns(runtime.NumCPU())
	}

	// Improve SQLite performance
	if _, err := sqlite.ExecContext(ctx, `
		pragma wal_autocheckpoint = 0;
		pragma temp_store         = memory;
	`); err != nil {
		return nil, err
	}

	// Initialize database
	if _, err := sqlite.ExecContext(ctx, `
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
	increment, err := sqlite.PrepareContext(ctx, `
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
		) on conflict do update set
			"count" = "count" + 1
	`)
	if err != nil {
		return nil, err
	}
	mostFrequent, err := sqlite.PrepareContext(ctx, `
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

	return &db{
		ctx:          ctx,
		db:           sqlite,
		increment:    increment,
		mostFrequent: mostFrequent,
	}, nil
}

var (
	i  int
	m  sync.Mutex
	wg sync.WaitGroup
)

func (s *db) Increment(cfg fizzbuzz.Config) error {
	m.Lock()
	i++
	if i == 1000 {
		i = 0
		wg.Wait()
		var failed bool
		err := s.db.QueryRow(`pragma wal_checkpoint(restart)`).Scan(&failed, new(int), new(int))
		if err == nil && failed {
			err = errors.New("pragma wal_checkpoint(restart) failed")
		}
		if err != nil {
			m.Unlock()
			return err
		}
	}
	m.Unlock()
	wg.Add(1)
	defer wg.Done()

	_, err := s.increment.ExecContext(s.ctx,
		cfg.Limit,
		cfg.Int1,
		cfg.Int2,
		cfg.Str1,
		cfg.Str2,
	)
	return err
}

func (s *db) MostFrequent() (count int, cfg fizzbuzz.Config, err error) {
	err = s.mostFrequent.QueryRowContext(s.ctx).Scan(
		&cfg.Limit,
		&cfg.Int1,
		&cfg.Int2,
		&cfg.Str1,
		&cfg.Str2,
		&count,
	)
	if err == sql.ErrNoRows {
		return 0, cfg, nil
	}
	return
}

func (s *db) Close() error {
	stmts := []*sql.Stmt{
		s.increment,
		s.mostFrequent,
	}
	for _, stmt := range stmts {
		if err := stmt.Close(); err != nil {
			return err
		}
	}
	return s.db.Close()
}
