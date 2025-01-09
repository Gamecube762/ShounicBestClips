package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func LoadDatabase(file string) (db *Database, err error) {
	err = os.MkdirAll("data", 0750)
	if err != nil {
		return
	}
	dbFile := filepath.Join("data", file)

	// https://github.com/mattn/go-sqlite3/issues/1179#issuecomment-1638083995
	// dedicated read & write connection
	// we can have unlimited reads, but at-most one write
	var cpu = max(4, runtime.NumCPU())
	fmt.Printf("Starting ReadConnection with %d connection(s)\n", cpu)
	readConn, err := sql.Open("sqlite3", dbFile)
	readConn.SetMaxOpenConns(cpu)
	if err != nil {
		return // panic
	}

	fmt.Printf("Starting WriteConnection\n")
	writeConn, err := sql.Open("sqlite3", dbFile)
	writeConn.SetMaxOpenConns(1)

	db = &Database{
		readPool:  readConn,
		writePool: writeConn,
	}

	if err == nil {
		err = db.setup()
	}

	return
}

type Database struct {
	readPool  *sql.DB
	writePool *sql.DB
}

func (db *Database) Close() {
	fmt.Println()

	fmt.Println("Closing ReadPool")
	err := db.readPool.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	fmt.Println("Closing WritePool")
	err = db.writePool.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println("DB sealed")
}

const db_pragmas = `
	PRAGMA journal_mode = WAL;
	PRAGMA busy_timeout = 5000;
	PRAGMA synchronous = NORMAL;
	PRAGMA cache_size = 1000000000;
	PRAGMA foreign_keys = true;
	PRAGMA temp_store = memory;
`

// the sql schema
const schema = `
	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY NOT NULL,
		url TEXT UNIQUE NOT NULL
	) STRICT;

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY NOT NULL,
		ip TEXT UNIQUE NOT NULL
	) STRICT;

	CREATE TABLE IF NOT EXISTS votes (
		user_id INTEGER NOT NULL,
		video_url TEXT NOT NULL,
		score INTEGER NOT NULL
	) STRICT;

	CREATE TABLE IF NOT EXISTS active_votes (
		user_id INTEGER PRIMARY KEY NOT NULL,
		start_time INTEGER NOT NULL,
		a TEXT, 
		b TEXT 
	) STRICT;
`

// Setup the database.
// Ran every time we load the database.
func (db *Database) setup() (err error) {
	// pragma (configs) first
	db.writePool.Exec(db_pragmas)
	db.readPool.Exec(db_pragmas)

	// load up the schemas, should be safe
	db.writePool.Exec(schema)

	const dummyVideos = `
		INSERT OR IGNORE INTO videos (url) 
			VALUES 
			('https://www.youtube.com/embed/oxEUk5c1iGU'),
			('https://www.youtube.com/embed/TFNYbCGCIaw'),
			('https://www.youtube.com/embed/N0qzSv9c0IY'),
			('https://www.youtube.com/embed/sZN5yJqDaYI'),
			('https://www.youtube.com/embed/TQllQlElpz8'),
			('https://www.youtube.com/embed/WRRC-Iw_OPg'),
			('https://www.youtube.com/embed/72eGw4H2Ka8'),
			('https://www.youtube.com/embed/4LilrtDfLP0'),
			('https://www.youtube.com/embed/uSlB4eznXoA'),
			('https://www.youtube.com/embed/i9bYnBb42oY'),
			('https://www.youtube.com/embed/lNfCvZl3sKw'),
			('https://www.youtube.com/embed/nz_BY7X44kc'),
			('https://www.youtube.com/embed/xrziHnudx3g'),
			('https://www.youtube.com/embed/2WNrx2jq184'),
			('https://www.youtube.com/embed/el0jsvcOSTg'),
			('https://www.youtube.com/embed/4hpbK7V146A'),
			('https://www.youtube.com/embed/Ta_-UPND0_M'),
			('https://www.youtube.com/embed/JgJUbmGDc6k'),
			('https://www.youtube.com/embed/ttArr90NvWo'),
			('https://www.youtube.com/embed/mIpnpYsl-VY'),
			('https://www.youtube.com/embed/4LilrtDfLP0'),
			('https://www.youtube.com/embed/0pnwE_Oy5WI');
	`

	// then load dummy data
	_, err = db.writePool.Exec(dummyVideos)
	return
}

func (db *Database) GetUser(remoteAddr string) (user User, err error) {
	user.ip = remoteAddr

	// Get user from database
	err = db.readPool.QueryRow(
		"SELECT id FROM users WHERE ip=?",
		user.ip,
	).Scan(&user.id)

	if err == sql.ErrNoRows {
		// Add user if they do not already exist.
		// Since this is doing INSERT it is a write operation
		err = db.writePool.QueryRow(
			"INSERT INTO users(ip) VALUES (?) RETURNING id",
			user.ip,
		).Scan(&user.id)
		return
	}
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
	_, err = db.writePool.Exec(
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
	row, err := db.readPool.Query(
		"SELECT url FROM videos WHERE url NOT IN (SELECT video_url FROM votes WHERE user_id = ?) ORDER BY random() LIMIT 2",
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
	row, err := db.readPool.Query(
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

	var other string
	switch choice {
	case vote.A:
		other = vote.B
	case vote.B:
		other = vote.A
	default:
		fmt.Println("Invalid choice")
		return
	}

	// TODO only supports one round of votes
	_, err = db.writePool.Exec(
		"DELETE FROM active_votes WHERE user_id = ?;"+
			"INSERT INTO votes VALUES (?, ?, 1), (?, ?, 0);",
		user.id,
		user.id,
		choice,
		user.id,
		other,
	)
	return
}

func (db *Database) TallyVotes() (count map[string]int, err error) {
	count = make(map[string]int)

	// Populate the map
	row, err := db.readPool.Query("SELECT url FROM videos")
	if err != nil {
		return
	}
	defer row.Close()
	for row.Next() {
		var url string
		err = row.Scan(&url)
		if err != nil {
			return
		}
		count[url] = 0
	}

	// Count the results
	row2, err := db.readPool.Query("SELECT video_url, score FROM votes")
	if err != nil {
		return
	}
	defer row2.Close()
	for row2.Next() {
		var url string
		var score int
		err = row2.Scan(&url, &score)
		if err != nil {
			return
		}
		count[url] += score
	}

	return
}
