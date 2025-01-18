package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/mattn/go-sqlite3"
)

func LoadDatabase(file string) (db *Database, err error) {
	err = os.MkdirAll("data", 0750)
	if err != nil {
		return
	}

	conn, err := sql.Open("sqlite3", filepath.Join("data", file))
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
		"CREATE TABLE IF NOT EXISTS videos (id INTEGER PRIMARY KEY, url TEXT UNIQUE, uploader_username TEXT)",
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, ip TEXT UNIQUE)",
		"CREATE TABLE IF NOT EXISTS culled_videos (url TEXT UNIQUE)", // TODO should be using Ids

		// TODO constraint for (user, video) pairs
		"CREATE TABLE IF NOT EXISTS votes (user_id INTEGER NOT NULL, video_url TEXT NOT NULL, score INTEGER NOT NULL, vote_time INTEGER NOT NULL)",
		// "CREATE UNIQUE IF NOT EXISTS INDEX idx_votes_user_video ON votes(user_id, video_url)",

		"CREATE TABLE IF NOT EXISTS active_votes ( " +
			"user_id INTEGER PRIMARY KEY NOT NULL, " +
			"start_time INTEGER NOT NULL, " +
			"a TEXT, " +
			"b TEXT " +
			")",

		// TEST
		// Now with family guy clips so the test stuff is usable
		//"INSERT OR IGNORE INTO videos (url) VALUES " +
		//	"('bAK0-WCjwiQ?si=jIJPZMtG6QF8i8k9')," +
		//	"('7OXGCi1sgkI')," +
		//	"('OjHOAfHokqk?si=7HZ2r-UvP5wxtWDy&amp;clip=UgkxhE5qzHs_XWtmAYDJ2QF8YTMnMG6DesNq&amp;clipt=EMf5FRjP5xc')," +
		//	"('rPle3Y4FLEg?si=OFcIVH1WbUcz5b5j')," +
		//	"('OjHOAfHokqk?si=I9gYK1eXCnd0pkmF')," +
		//	"('WRRC-Iw_OPg')," +
		//	"('72eGw4H2Ka8')," +
		//	"('4LilrtDfLP0')," +
		//	"('uSlB4eznXoA')," +
		//	"('i9bYnBb42oY')," +
		//	"('lNfCvZl3sKw')," +
		//	"('nz_BY7X44kc')," +
		//	"('xrziHnudx3g')," +
		//	"('4hpbK7V146A')," +
		//	"('Ta_-UPND0_M')," +
		//	"('JgJUbmGDc6k')," +
		//	"('ttArr90NvWo')," +
		//	"('mIpnpYsl-VY')," +
		//	"('4LilrtDfLP0')," +
		//	"('duAGuYeF7zY')," +
		//	"('0pnwE_Oy5WI')",
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
	user.ip = remoteAddr

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

	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "debug",
		Category: "users",
		Message:  "Adding new user",
	})

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
		time.Now().Unix(),
		a,
		b,
	)
	return
}

// Get new vote options for the user
// Empty a or b strings means not enough available voting options
func (db *Database) findNextPair(user User) (a string, b string, err error) {
	row, err := db.Query(
		"SELECT url FROM videos "+
			//"WHERE url NOT IN (SELECT video_url FROM votes WHERE user_id = ?) "+
			"WHERE   url NOT IN (SELECT url FROM culled_videos) "+
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
	if vote.startTime.Add(voteCooldown * time.Second).After(time.Now()) {
		// User voting too fast, ignore vote
		return fmt.Errorf("too fast")
	}

	// TODO limit max time? 12hours?

	var other string
	switch choice {
	case vote.A:
		other = vote.B
	case vote.B:
		other = vote.A
	default:
		fmt.Printf("Invalid choice \"%s\"\n", choice)
		return
	}

	// TODO only supports one round of votes
	_, err = db.Exec(
		"DELETE FROM active_votes WHERE user_id = ?;"+
			"INSERT INTO votes VALUES (?, ?, 1, ?), (?, ?, 0, ?);",
		user.id,
		user.id,
		choice,
		time.Now().UnixMilli(),
		user.id,
		other,
		time.Now().UnixMilli(),
	)
	return
}

func (db *Database) TallyVotes() (count map[string]int, err error) {
	count = make(map[string]int)

	// Populate the map
	row, err := db.Query("SELECT url FROM videos")
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
	row2, err := db.Query("SELECT video_url, score FROM votes")
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

func (db *Database) GetTotalClips() (grandTotal int64) {
	row, err := db.Query("SELECT (SELECT COUNT() FROM videos) as totalClips, (SELECT COUNT() FROM culled_videos) as totalCulled")
	if err != nil {
		return 0
	}
	defer row.Close()

	if !row.Next() {
		// 0 videos available
		return 0
	}

	var totalClips int64
	var totalCulled int64
	err = row.Scan(&totalClips, &totalCulled)
	grandTotal = totalClips - totalCulled

	if err != nil {
		return 0
	}

	return
}
