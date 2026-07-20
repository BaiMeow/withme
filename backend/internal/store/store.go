// Package store 负责相亲资料的持久化，兼容 sqlite（本地）和 mysql（线上）。
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	"withme/internal/model"
)

// ErrNotFound 分享 ID 不存在
var ErrNotFound = errors.New("profile not found")

type Store struct {
	db     *sql.DB
	driver string
}

// Open 按 driver 打开数据库并自动建表。driver 支持 sqlite / mysql。
func Open(driver, dsn string) (*Store, error) {
	var sqlDriver string
	switch driver {
	case "sqlite":
		sqlDriver = "sqlite"
		if !strings.Contains(dsn, "?") {
			if dir := filepath.Dir(dsn); dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return nil, fmt.Errorf("create sqlite dir: %w", err)
				}
			}
			dsn += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
		}
	case "mysql":
		sqlDriver = "mysql"
		if !strings.Contains(dsn, "parseTime") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "parseTime=true"
		}
	default:
		return nil, fmt.Errorf("unsupported database driver %q (want sqlite or mysql)", driver)
	}

	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", driver, err)
	}
	// 单写多读场景，避免 sqlite 数据库文件锁
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping %s: %w", driver, err)
	}

	s := &Store{db: db, driver: driver}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	var ddl string
	if s.driver == "mysql" {
		ddl = `CREATE TABLE IF NOT EXISTS profiles (
			id VARCHAR(16) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			version VARCHAR(16) NOT NULL,
			profile JSON NOT NULL,
			views BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	} else {
		ddl = `CREATE TABLE IF NOT EXISTS profiles (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			version TEXT NOT NULL,
			profile TEXT NOT NULL,
			views INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_profiles_created_at ON profiles(created_at)`
	}
	for _, stmt := range strings.Split(ddl, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Save 保存一份资料并返回分享 ID。
func (s *Store) Save(ctx context.Context, username, version string, p *model.DatingProfile) (string, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("marshal profile: %w", err)
	}
	id, err := newShareID()
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO profiles (id, username, version, profile, views, created_at) VALUES (?, ?, ?, ?, 0, ?)`,
		id, username, version, string(raw), time.Now())
	if err != nil {
		return "", fmt.Errorf("insert profile: %w", err)
	}
	return id, nil
}

// Get 按分享 ID 取出资料，并将浏览次数 +1。
func (s *Store) Get(ctx context.Context, id string) (*model.StoredProfile, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE profiles SET views = views + 1 WHERE id = ?`, id); err != nil {
		return nil, fmt.Errorf("bump views: %w", err)
	}
	return s.get(ctx, id)
}

// Peek 按分享 ID 取出资料，不增加浏览次数。
func (s *Store) Peek(ctx context.Context, id string) (*model.StoredProfile, error) {
	return s.get(ctx, id)
}

func (s *Store) get(ctx context.Context, id string) (*model.StoredProfile, error) {
	var sp model.StoredProfile
	var raw string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, version, profile, views, created_at FROM profiles WHERE id = ?`, id).
		Scan(&sp.ID, &sp.Username, &sp.Version, &raw, &sp.Views, &sp.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query profile: %w", err)
	}
	var p model.DatingProfile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}
	sp.Profile = &p
	return &sp, nil
}

// ListRecent 返回最近生成的资料摘要（不含正文）。
func (s *Store) ListRecent(ctx context.Context, limit int) ([]model.ProfileSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, username, version, profile, views, created_at FROM profiles ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	out := make([]model.ProfileSummary, 0)
	for rows.Next() {
		var sum model.ProfileSummary
		var raw string
		if err := rows.Scan(&sum.ID, &sum.Username, &sum.Version, &raw, &sum.Views, &sum.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		var p model.DatingProfile
		if err := json.Unmarshal([]byte(raw), &p); err == nil {
			sum.Nickname = p.Nickname
			sum.Occupation = p.BasicInfo.Occupation
		}
		out = append(out, sum)
	}
	return out, rows.Err()
}

// newShareID 生成 8 位 base62 分享 ID。
func newShareID() (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate share id: %w", err)
	}
	for i, b := range buf {
		buf[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(buf), nil
}
