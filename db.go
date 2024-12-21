package main

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func LoadDatabase(file string) (db *Database, err error) {
	conn, err := sql.Open("sqlite3", file)
	db = &Database{conn}

	if err == nil {
		err = db.setup()
	}

	return
}

type Database struct{ *sql.DB }

// Setup the database.
// Ran every time we load the database.
func (db *Database) setup() (err error) {
	var setupQueries = []string{
		// TODO ? video: title, uploader, docSubmitter, upload date
		"CREATE TABLE IF NOT EXISTS videos (id INTEGER PRIMARY KEY, url TEXT UNIQUE)",
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, ip TEXT UNIQUE)",

		// TODO constraint for (user, video) pairs
		"CREATE TABLE IF NOT EXISTS votes (user_id INTEGER NOT NULL, video_url TEXT NOT NULL, score INTEGER NOT NULL)",
		// "CREATE UNIQUE IF NOT EXISTS INDEX idx_votes_user_video ON votes(user_id, video_url)",

		"CREATE TABLE IF NOT EXISTS active_votes ( " +
			"user_id INTEGER PRIMARY KEY NOT NULL, " +
			"start_time INTEGER NOT NULL, " +
			"a TEXT, " +
			"b TEXT " +
			")",

		// TEST
		// Now with family guy clips so the test stuff is usable
		"INSERT OR IGNORE INTO videos (url) VALUES " +
			"('https://www.youtube.com/embed/JgJUbmGDc6k')," +
			"('https://www.youtube.com/embed/ttArr90NvWo')," +
			"('https://www.youtube.com/embed/mIpnpYsl-VY')," +
			"('https://www.youtube.com/embed/4LilrtDfLP0')," +
			"('https://www.youtube.com/embed/0pnwE_Oy5WI')",
	}

	// Transaction so we can undo if we error
	tran, err := db.Begin()
	if err != nil {
		return
	}

	// Run all setupQueries
	for _, query := range setupQueries {
		_, err = db.Exec(query)
		if err != nil {
			tran.Rollback()
			return
		}
	}

	// Commit transaction
	return tran.Commit()
}

func (db *Database) GetUser(remoteAddr string) (user User, err error) {
	user.ip, _, err = net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}

	// Get user from database
	row, err := db.Query(
		"SELECT id FROM users WHERE ip=?",
		user.ip,
	)
	if err != nil {
		return
	}
	defer row.Close()

	if row.Next() {
		err = row.Scan(&user.id)
		return
	}

	// Add user if they do not already exist.
	row2, err := db.Query(
		"INSERT INTO users(ip) VALUES (?) RETURNING id",
		user.ip,
	)
	if err != nil {
		return
	}
	defer row2.Close()

	if !row2.Next() {
		// Database has dementia
		err = fmt.Errorf("GetUser gave 0 results after inserting %s", user.ip)
		return
	}

	err = row2.Scan(&user.id)
	return
}

// Get the next vote for a user
// If a vote already exists, it will be deleted.
// If there are < 2 options, `vote` will be nil
func (db *Database) GetNextVoteForUser(user User) (vote *VoteOptions, err error) {
	a, b, err := db.findNextPair(user)
	if a == "" || b == "" || err != nil {
		// Return nil vote, we don't have enough
		// voting options for this user
		return
	}

	// Exec is 10x-100x slower for some reason.
	// Query has issues committing inserts
	// Locking issue?
	vote = &VoteOptions{time.Now(), a, b}
	_, err = db.Exec(
		"INSERT OR REPLACE INTO active_votes VALUES (?, ?, ?, ?)",
		user.id,
		time.Now().UnixMilli(),
		a,
		b,
	)
	return
}

// Get new vote options for the user
// Empty a or b strings means not enough available voting options
func (db *Database) findNextPair(user User) (a string, b string, err error) {
	row, err := db.Query(
		"SELECT url FROM videos LEFT JOIN votes ON videos.url = votes.video_url "+
			"WHERE (votes.user_id == ? OR votes.user_id IS NULL) AND (votes.score IS NULL OR votes.score == 0) "+ // TODO vote.score == ? for round X of voting
			"ORDER BY random() LIMIT 2",
		user.id,
	)
	if err != nil {
		return
	}
	defer row.Close()

	if !row.Next() {
		// 0 videos available
		return
	}

	err = row.Scan(&a)
	if err != nil || !row.Next() {
		return
	}

	err = row.Scan(&b)
	return
}

func (db *Database) GetCurrentVotingOptionsForUser(user User) (vote *VoteOptions, err error) {
	row, err := db.Query(
		"SELECT start_time, a, b FROM active_votes WHERE user_id = ?",
		user.id,
	)
	if err != nil {
		return
	}
	defer row.Close()

	if !row.Next() {
		// User has no vote options, returning nil
		return
	}

	var startTime int64
	vote = &VoteOptions{}
	err = row.Scan(
		&startTime,
		&vote.A,
		&vote.B,
	)

	vote.startTime = time.Unix(startTime, 0)

	return
}

func (db *Database) SubmitUserVote(user User, choice string) (err error) {
	vote, err := db.GetCurrentVotingOptionsForUser(user)
	if err != nil || vote == nil {
		// If the user has no options, we'll do nothing
		return
	}

	// TODO scale min time to video length
	// 	?	minTime := max(min(a.length, b.length) / 2, 90 * time.seconds)
	// if vote.startTime.Add(30 * time.Second).After(time.Now()) {
	// 	// User voting too fast, ignore vote
	// 	return fmt.Errorf("too fast")
	// }

	// TODO limit max time? 12hours?

	if choice != vote.A && choice != vote.B {
		fmt.Println("Invalid choice")
		return
	}

	// TODO only supports one round of votes
	_, err = db.Exec(
		"DELETE FROM active_votes WHERE user_id = ?;"+
			"INSERT INTO votes VALUES (?, ?, 1);",
		user.id,
		user.id,
		choice,
	)
	return
}
