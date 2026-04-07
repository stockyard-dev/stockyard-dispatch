package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "dispatch.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	conn.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL,record_id TEXT NOT NULL,data TEXT NOT NULL DEFAULT '{}',PRIMARY KEY(resource, record_id))`)
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS lists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS subscribers (
    id TEXT PRIMARY KEY,
    list_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT DEFAULT '',
    status TEXT DEFAULT 'active',
    token TEXT NOT NULL,
    subscribed_at TEXT DEFAULT (datetime('now')),
    unsubscribed_at TEXT DEFAULT '',
    UNIQUE(list_id, email)
);
CREATE INDEX IF NOT EXISTS idx_subs_list ON subscribers(list_id);
CREATE INDEX IF NOT EXISTS idx_subs_email ON subscribers(email);
CREATE INDEX IF NOT EXISTS idx_subs_token ON subscribers(token);

CREATE TABLE IF NOT EXISTS campaigns (
    id TEXT PRIMARY KEY,
    list_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_html TEXT DEFAULT '',
    body_text TEXT DEFAULT '',
    status TEXT DEFAULT 'draft',
    sent_count INTEGER DEFAULT 0,
    open_count INTEGER DEFAULT 0,
    click_count INTEGER DEFAULT 0,
    sent_at TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_camp_list ON campaigns(list_id);

CREATE TABLE IF NOT EXISTS sends (
    id TEXT PRIMARY KEY,
    campaign_id TEXT NOT NULL,
    subscriber_id TEXT NOT NULL,
    email TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    opened INTEGER DEFAULT 0,
    clicked INTEGER DEFAULT 0,
    sent_at TEXT DEFAULT '',
    error TEXT DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sends_camp ON sends(campaign_id);
CREATE INDEX IF NOT EXISTS idx_sends_sub ON sends(subscriber_id);
`)
	return err
}

// --- Lists ---

type List struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	SubCount    int    `json:"subscriber_count"`
}

func (db *DB) CreateList(name, desc string) (*List, error) {
	id := "lst_" + genID(6)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO lists (id,name,description,created_at) VALUES (?,?,?,?)", id, name, desc, now)
	if err != nil {
		return nil, err
	}
	return &List{ID: id, Name: name, Description: desc, CreatedAt: now}, nil
}

func (db *DB) ListLists() ([]List, error) {
	rows, err := db.conn.Query(`SELECT l.id, l.name, l.description, l.created_at,
		(SELECT COUNT(*) FROM subscribers WHERE list_id=l.id AND status='active')
		FROM lists l ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []List
	for rows.Next() {
		var l List
		rows.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.SubCount)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (db *DB) GetList(id string) (*List, error) {
	var l List
	err := db.conn.QueryRow(`SELECT l.id, l.name, l.description, l.created_at,
		(SELECT COUNT(*) FROM subscribers WHERE list_id=l.id AND status='active')
		FROM lists l WHERE l.id=?`, id).
		Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.SubCount)
	return &l, err
}

func (db *DB) DeleteList(id string) error {
	db.conn.Exec("DELETE FROM subscribers WHERE list_id=?", id)
	db.conn.Exec("DELETE FROM campaigns WHERE list_id=?", id)
	_, err := db.conn.Exec("DELETE FROM lists WHERE id=?", id)
	return err
}

// --- Subscribers ---

type Subscriber struct {
	ID             string `json:"id"`
	ListID         string `json:"list_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Token          string `json:"token,omitempty"`
	SubscribedAt   string `json:"subscribed_at"`
	UnsubscribedAt string `json:"unsubscribed_at,omitempty"`
}

func (db *DB) AddSubscriber(listID, email, name string) (*Subscriber, error) {
	id := "sub_" + genID(8)
	token := genID(16)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO subscribers (id,list_id,email,name,token,subscribed_at) VALUES (?,?,?,?,?,?)",
		id, listID, email, name, token, now)
	if err != nil {
		return nil, err
	}
	return &Subscriber{ID: id, ListID: listID, Email: email, Name: name, Status: "active", Token: token, SubscribedAt: now}, nil
}

func (db *DB) ListSubscribers(listID string, limit int) ([]Subscriber, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := db.conn.Query("SELECT id,list_id,email,name,status,subscribed_at,unsubscribed_at FROM subscribers WHERE list_id=? ORDER BY subscribed_at DESC LIMIT ?", listID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscriber
	for rows.Next() {
		var s Subscriber
		rows.Scan(&s.ID, &s.ListID, &s.Email, &s.Name, &s.Status, &s.SubscribedAt, &s.UnsubscribedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) ActiveSubscribers(listID string) ([]Subscriber, error) {
	rows, err := db.conn.Query("SELECT id,list_id,email,name,status,token,subscribed_at FROM subscribers WHERE list_id=? AND status='active'", listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subscriber
	for rows.Next() {
		var s Subscriber
		rows.Scan(&s.ID, &s.ListID, &s.Email, &s.Name, &s.Status, &s.Token, &s.SubscribedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) Unsubscribe(token string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("UPDATE subscribers SET status='unsubscribed', unsubscribed_at=? WHERE token=?", now, token)
	return err
}

func (db *DB) DeleteSubscriber(id string) error {
	_, err := db.conn.Exec("DELETE FROM subscribers WHERE id=?", id)
	return err
}

func (db *DB) TotalSubscribers() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM subscribers WHERE status='active'").Scan(&count)
	return count
}

// --- Campaigns ---

type Campaign struct {
	ID         string `json:"id"`
	ListID     string `json:"list_id"`
	Subject    string `json:"subject"`
	BodyHTML   string `json:"body_html,omitempty"`
	BodyText   string `json:"body_text,omitempty"`
	Status     string `json:"status"`
	SentCount  int    `json:"sent_count"`
	OpenCount  int    `json:"open_count"`
	ClickCount int    `json:"click_count"`
	SentAt     string `json:"sent_at,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func (db *DB) CreateCampaign(listID, subject, bodyHTML, bodyText string) (*Campaign, error) {
	id := "cmp_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO campaigns (id,list_id,subject,body_html,body_text,created_at) VALUES (?,?,?,?,?,?)",
		id, listID, subject, bodyHTML, bodyText, now)
	if err != nil {
		return nil, err
	}
	return &Campaign{ID: id, ListID: listID, Subject: subject, BodyHTML: bodyHTML, BodyText: bodyText,
		Status: "draft", CreatedAt: now}, nil
}

func (db *DB) ListCampaigns(listID string) ([]Campaign, error) {
	rows, err := db.conn.Query("SELECT id,list_id,subject,status,sent_count,open_count,click_count,sent_at,created_at FROM campaigns WHERE list_id=? ORDER BY created_at DESC", listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Campaign
	for rows.Next() {
		var c Campaign
		rows.Scan(&c.ID, &c.ListID, &c.Subject, &c.Status, &c.SentCount, &c.OpenCount, &c.ClickCount, &c.SentAt, &c.CreatedAt)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (db *DB) GetCampaign(id string) (*Campaign, error) {
	var c Campaign
	err := db.conn.QueryRow("SELECT id,list_id,subject,body_html,body_text,status,sent_count,open_count,click_count,sent_at,created_at FROM campaigns WHERE id=?", id).
		Scan(&c.ID, &c.ListID, &c.Subject, &c.BodyHTML, &c.BodyText, &c.Status, &c.SentCount, &c.OpenCount, &c.ClickCount, &c.SentAt, &c.CreatedAt)
	return &c, err
}

func (db *DB) UpdateCampaignStatus(id, status string) {
	now := time.Now().UTC().Format(time.RFC3339)
	if status == "sent" {
		db.conn.Exec("UPDATE campaigns SET status=?, sent_at=? WHERE id=?", status, now, id)
	} else {
		db.conn.Exec("UPDATE campaigns SET status=? WHERE id=?", status, id)
	}
}

func (db *DB) DeleteCampaign(id string) error {
	db.conn.Exec("DELETE FROM sends WHERE campaign_id=?", id)
	_, err := db.conn.Exec("DELETE FROM campaigns WHERE id=?", id)
	return err
}

// --- Sends ---

type Send struct {
	ID           string `json:"id"`
	CampaignID   string `json:"campaign_id"`
	SubscriberID string `json:"subscriber_id"`
	Email        string `json:"email"`
	Status       string `json:"status"`
	Opened       int    `json:"opened"`
	Clicked      int    `json:"clicked"`
	SentAt       string `json:"sent_at,omitempty"`
	Error        string `json:"error,omitempty"`
}

func (db *DB) CreateSend(campaignID, subscriberID, email string) (*Send, error) {
	id := "snd_" + genID(8)
	_, err := db.conn.Exec("INSERT INTO sends (id,campaign_id,subscriber_id,email) VALUES (?,?,?,?)",
		id, campaignID, subscriberID, email)
	if err != nil {
		return nil, err
	}
	return &Send{ID: id, CampaignID: campaignID, SubscriberID: subscriberID, Email: email, Status: "pending"}, nil
}

func (db *DB) UpdateSendStatus(id, status, errMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec("UPDATE sends SET status=?, sent_at=?, error=? WHERE id=?", status, now, errMsg, id)
}

func (db *DB) IncrementCampaignSent(id string) {
	db.conn.Exec("UPDATE campaigns SET sent_count=sent_count+1 WHERE id=?", id)
}

func (db *DB) RecordOpen(campaignID string) {
	db.conn.Exec("UPDATE campaigns SET open_count=open_count+1 WHERE id=?", campaignID)
}

func (db *DB) ListSends(campaignID string) ([]Send, error) {
	rows, err := db.conn.Query("SELECT id,campaign_id,subscriber_id,email,status,opened,clicked,sent_at,error FROM sends WHERE campaign_id=? ORDER BY sent_at DESC", campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Send
	for rows.Next() {
		var s Send
		rows.Scan(&s.ID, &s.CampaignID, &s.SubscriberID, &s.Email, &s.Status, &s.Opened, &s.Clicked, &s.SentAt, &s.Error)
		out = append(out, s)
	}
	return out, rows.Err()
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var lists, subs, campaigns, sent int
	db.conn.QueryRow("SELECT COUNT(*) FROM lists").Scan(&lists)
	db.conn.QueryRow("SELECT COUNT(*) FROM subscribers WHERE status='active'").Scan(&subs)
	db.conn.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&campaigns)
	db.conn.QueryRow("SELECT COUNT(*) FROM sends WHERE status='sent'").Scan(&sent)
	return map[string]any{"lists": lists, "subscribers": subs, "campaigns": campaigns, "emails_sent": sent}
}

func (db *DB) Cleanup(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec("DELETE FROM sends WHERE sent_at < ? AND sent_at != ''", cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.conn.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.conn.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.conn.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.conn.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
