package stats

import (
	"context"
	"database/sql"
	"net/url"
	"runtime"
	"strings"

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
	db := &db{ctx: ctx}
	var err error

	db.db, err = sql.Open("sqlite3", dataSourceName+"?"+url.Values{
		"_busy_timeout":        {"5000"},
		"_foreign_keys":        {"true"},
		"_journal_mode":        {"wal"},
		"_synchronous":         {"normal"},
		"_case_sensitive_like": {"true"},
		"_cache_size":          {"10000"},
	}.Encode())
	if err != nil {
		return nil, err
	}

	// Adjust database/sql settings to SQLite to avoid ever-growing WAL file
	if strings.Contains(dataSourceName, "memory") {
		db.db.SetMaxOpenConns(1)
	} else {
		db.db.SetMaxOpenConns(runtime.NumCPU())
		db.db.SetMaxIdleConns(runtime.NumCPU())
	}

	// Improve SQLite performance
	if _, err := db.db.ExecContext(ctx, `pragma temp_store = memory`); err != nil {
		return nil, err
	}

	// Initialize database
	if _, err := db.db.ExecContext(ctx, `
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
	db.increment, err = db.db.PrepareContext(ctx, `
		insert into "stat" (
			"limit",
			"int1",
			"int2",
			"str1",
			"str2",
			"count"
		) values (
			?, -- limit
			?, -- int1
			?, -- int2
			?, -- str1
			?, -- str2
			1  -- count
		) on conflict do update set
			"count" = "count" + 1;
	`)
	if err != nil {
		return nil, err
	}
	db.mostFrequent, err = db.db.PrepareContext(ctx, `
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

	return db, nil
}

func (s *db) Increment(cfg fizzbuzz.Config) error {
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
