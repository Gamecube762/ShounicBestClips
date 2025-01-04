package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"

	_ "modernc.org/sqlite"
)

func LoadDatabase(file string) (db *Database, err error) {
	conn, err := sql.Open("sqlite", file)
	db = &Database{conn}

	if err == nil {
		err = db.Setup()
		if err == nil {
			fmt.Println("Database loaded successfully")
		}
	}

	return
}

type Database struct{ *sql.DB }

// Setup sets up the database.
// Ran every time we load the database.
func (db *Database) Setup() (err error) {
	var setupQueries = []string{
		// TODO ? video: title, uploader, docSubmitter, upload date
		"CREATE TABLE IF NOT EXISTS videos (url TEXT NOT NULL PRIMARY KEY)",
		"CREATE TABLE IF NOT EXISTS users (id BLOB NOT NULL PRIMARY KEY, ip TEXT UNIQUE)",
		// We probably shouldn't constrain the userid and url, instead we should constrain the comparisons, because
		// a user can get the same video in a different comparison
		"CREATE TABLE IF NOT EXISTS votes (user_id BLOB NOT NULL, video_url TEXT NOT NULL, score INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES users(id), FOREIGN KEY (video_url) REFERENCES videos(url))",
		// "CREATE UNIQUE IF NOT EXISTS INDEX idx_votes_user_video ON votes(user_id, video_url)",

		"CREATE TABLE IF NOT EXISTS active_votes (user_id BLOB NOT NULL, id BLOB NOT NULL PRIMARY KEY, a TEXT NOT NULL, b TEXT NOT NULL, FOREIGN KEY (user_id) REFERENCES users(id))",

		// Lmao all the shounic videos because funny
		"INSERT INTO videos ( url) VALUES " +
			"('https://www.youtube.com/embed/hKghMBaU43I')," +
			"('https://www.youtube.com/embed/ON2DukHNAFM')," +
			"('https://www.youtube.com/embed/0_lwkKtv1-A')," +
			"('https://www.youtube.com/embed/rUTwWhM89Qo')," +
			"('https://www.youtube.com/embed/jS9R32G2-lo')," +
			"('https://www.youtube.com/embed/GcfYEyt_3Ag')," +
			"('https://www.youtube.com/embed/7baMnAFDHVI')," +
			"('https://www.youtube.com/embed/sBJzNH0jkuA')," +
			"('https://www.youtube.com/embed/6C8O5xfxtfw')," +
			"('https://www.youtube.com/embed/Q1jAUWwtrAo')," +
			"('https://www.youtube.com/embed/a_gSPx_nDkI')," +
			"('https://www.youtube.com/embed/griZgfJukZs')," +
			"('https://www.youtube.com/embed/6hC4ccOEK08')," +
			"('https://www.youtube.com/embed/Z2eduTNisYA')," +
			"('https://www.youtube.com/embed/TvH7syo376E')," +
			"('https://www.youtube.com/embed/6D1Rc1dWEzs')," +
			"('https://www.youtube.com/embed/UMi847MdESY')," +
			"('https://www.youtube.com/embed/dkjDzQIyrj8')," +
			"('https://www.youtube.com/embed/aVeV-nRYn1U')," +
			"('https://www.youtube.com/embed/1y1P9PIYNHU')," +
			"('https://www.youtube.com/embed/8I082SOybkQ')," +
			"('https://www.youtube.com/embed/mWkpUD_l4AQ')," +
			"('https://www.youtube.com/embed/Pp97A1VjIOg')," +
			"('https://www.youtube.com/embed/znFmPQGTUIw')," +
			"('https://www.youtube.com/embed/z1gbDdhcUpI')," +
			"('https://www.youtube.com/embed/FpuqQ2IuYvc')," +
			"('https://www.youtube.com/embed/0I4EFNVHP0w')," +
			"('https://www.youtube.com/embed/oy96M7sXEc8')," +
			"('https://www.youtube.com/embed/M-RJ0AQsRTw')," +
			"('https://www.youtube.com/embed/uoS2lTZ3gtQ')," +
			"('https://www.youtube.com/embed/OIvbxikwEbU')," +
			"('https://www.youtube.com/embed/LKwQveusMk8')," +
			"('https://www.youtube.com/embed/7buYRDM_v0w')," +
			"('https://www.youtube.com/embed/JmnI-DF1TOI')," +
			"('https://www.youtube.com/embed/eFFM12AvXdQ')," +
			"('https://www.youtube.com/embed/TQINJNbQlb0')," +
			"('https://www.youtube.com/embed/87kvEH2jPIQ')," +
			"('https://www.youtube.com/embed/WbGo1KY24Ho')," +
			"('https://www.youtube.com/embed/xSklooh0rKc')," +
			"('https://www.youtube.com/embed/Eya0idPzv2A')," +
			"('https://www.youtube.com/embed/zOpMuPZlNdQ')," +
			"('https://www.youtube.com/embed/7A2CUnk1TQs')," +
			"('https://www.youtube.com/embed/5m7V9zWlYdM')," +
			"('https://www.youtube.com/embed/GFXGwko225k')," +
			"('https://www.youtube.com/embed/BvFjUQ2Kyi8')," +
			"('https://www.youtube.com/embed/PwTFA93W6G8')," +
			"('https://www.youtube.com/embed/PEhY4vE6krE')," +
			"('https://www.youtube.com/embed/s7GTiBs3hRw')," +
			"('https://www.youtube.com/embed/76HDJIWfVy4')," +
			"('https://www.youtube.com/embed/DiahjyxTSk0')," +
			"('https://www.youtube.com/embed/6yM6Fkl9SoE')," +
			"('https://www.youtube.com/embed/hCUh5cRCyK8')," +
			"('https://www.youtube.com/embed/hnWsTZoFrOw')," +
			"('https://www.youtube.com/embed/WUDSXjh4X2k')," +
			"('https://www.youtube.com/embed/hqFoKHE95Eg')," +
			"('https://www.youtube.com/embed/Nok264ZCxBg')," +
			"('https://www.youtube.com/embed/CtG6H1r8QmE')," +
			"('https://www.youtube.com/embed/4ZxnOG8-Hos')," +
			"('https://www.youtube.com/embed/x6nIev0elY4')," +
			"('https://www.youtube.com/embed/3Cz7Id_pIQA')," +
			"('https://www.youtube.com/embed/_l7aVVhxIzc')," +
			"('https://www.youtube.com/embed/7vbJ8Vu-XHU')," +
			"('https://www.youtube.com/embed/OoMxTs394oM')," +
			"('https://www.youtube.com/embed/E4xlvbLcrO0')," +
			"('https://www.youtube.com/embed/JsShwxBc9no')," +
			"('https://www.youtube.com/embed/C0CMW4qQCBk')," +
			"('https://www.youtube.com/embed/nrsDklFXrpc')," +
			"('https://www.youtube.com/embed/lN2ERuo9vsg')," +
			"('https://www.youtube.com/embed/vfj-oE94cgs')," +
			"('https://www.youtube.com/embed/VwEJUGwzm7U')," +
			"('https://www.youtube.com/embed/VIKnDQBiVv4')," +
			"('https://www.youtube.com/embed/7z_p_RqLhkA')," +
			"('https://www.youtube.com/embed/pw2X1yhrDdE')," +
			"('https://www.youtube.com/embed/P4RbXa_Twh8')," +
			"('https://www.youtube.com/embed/dPyMG23LTX0')," +
			"('https://www.youtube.com/embed/mlhw6RqvOgs')," +
			"('https://www.youtube.com/embed/NgHygsNwTNk')," +
			"('https://www.youtube.com/embed/jPKuyeDb0mM')," +
			"('https://www.youtube.com/embed/EO5zj-DwJoA')," +
			"('https://www.youtube.com/embed/EDsDnR2dzlw')," +
			"('https://www.youtube.com/embed/YFLgtu7Z3gM')," +
			"('https://www.youtube.com/embed/ljl1jBEY3_A')," +
			"('https://www.youtube.com/embed/jIwqlKDPq4s')," +
			"('https://www.youtube.com/embed/eEBr-YTQg5M')," +
			"('https://www.youtube.com/embed/TGulB0MfxPs')," +
			"('https://www.youtube.com/embed/SgkgsgaBBCA')," +
			"('https://www.youtube.com/embed/n8VJ-5UekWk')," +
			"('https://www.youtube.com/embed/ehlrUPrvFuk')," +
			"('https://www.youtube.com/embed/XmfW3Fhzc1s')," +
			"('https://www.youtube.com/embed/iwxbY-p_w0w')," +
			"('https://www.youtube.com/embed/QzZoo1yAQag')," +
			"('https://www.youtube.com/embed/oEM6lP7tUJA')," +
			"('https://www.youtube.com/embed/7m1U8rvamrQ')," +
			"('https://www.youtube.com/embed/KsiHWun1nkE')," +
			"('https://www.youtube.com/embed/EwUbXd5XUgA')," +
			"('https://www.youtube.com/embed/O0QfINYF8CE')," +
			"('https://www.youtube.com/embed/iIr47SNZK8g')," +
			"('https://www.youtube.com/embed/NxeljS6GE9g')," +
			"('https://www.youtube.com/embed/WHGVAJgHMX8')," +
			"('https://www.youtube.com/embed/w4nSnumKp88')," +
			"('https://www.youtube.com/embed/E4rDU-UVzeE')," +
			"('https://www.youtube.com/embed/vr5wP5ysBXw')," +
			"('https://www.youtube.com/embed/cxvOOryIlzc')," +
			"('https://www.youtube.com/embed/BzFi0iFoLoU')," +
			"('https://www.youtube.com/embed/ytpasNctt8I')," +
			"('https://www.youtube.com/embed/W3BMzt-4s3E')," +
			"('https://www.youtube.com/embed/OjHOAfHokqk')," +
			"('https://www.youtube.com/embed/hHZdmn3U32c')," +
			"('https://www.youtube.com/embed/8iEXhbqami8')," +
			"('https://www.youtube.com/embed/8RjDM6-nLxc')," +
			"('https://www.youtube.com/embed/jvOaUF1Sl5c')," +
			"('https://www.youtube.com/embed/UFtZMIWt0WI')," +
			"('https://www.youtube.com/embed/CoQYWDsDH94')," +
			"('https://www.youtube.com/embed/AUPBC5W1KHo')," +
			"('https://www.youtube.com/embed/VkA7cyAZDts')," +
			"('https://www.youtube.com/embed/ntaxvBPVadY')," +
			"('https://www.youtube.com/embed/67LPSFtVlsk')," +
			"('https://www.youtube.com/embed/WLx_3bON0Mw')," +
			"('https://www.youtube.com/embed/BAs-ph7lqqA')," +
			"('https://www.youtube.com/embed/DXNM-EPPVcQ')," +
			"('https://www.youtube.com/embed/8BUVGg175_M')," +
			"('https://www.youtube.com/embed/YdcDQjuEVII')," +
			"('https://www.youtube.com/embed/-CgSTV36ZZw')," +
			"('https://www.youtube.com/embed/8TfpAmiisQI')," +
			"('https://www.youtube.com/embed/D06tRhV1gYE')," +
			"('https://www.youtube.com/embed/YURn9IkYSqI')," +
			"('https://www.youtube.com/embed/TJM0AiMhqDg')," +
			"('https://www.youtube.com/embed/3jUPs_hIhSE')," +
			"('https://www.youtube.com/embed/zwz5yJR_aFA')," +
			"('https://www.youtube.com/embed/hxMKP5hyzB8')," +
			"('https://www.youtube.com/embed/bDlgOUOJqWk')," +
			"('https://www.youtube.com/embed/A3dBYoFUF-g')," +
			"('https://www.youtube.com/embed/-S7RHybAwGU')," +
			"('https://www.youtube.com/embed/1kPIz0DfPcY')," +
			"('https://www.youtube.com/embed/RdTJHVG_IdU')," +
			"('https://www.youtube.com/embed/Cu3Anpl3Se0')," +
			"('https://www.youtube.com/embed/dk0Ue7R69iM')," +
			"('https://www.youtube.com/embed/LAARoaRBrlg')," +
			"('https://www.youtube.com/embed/xe0P0rnsS1Q')," +
			"('https://www.youtube.com/embed/Q5Jj-wnl9w4')," +
			"('https://www.youtube.com/embed/GSfnxqlj2T4')," +
			"('https://www.youtube.com/embed/iVthhSshJp0')," +
			"('https://www.youtube.com/embed/E7oYZAzsNrA')," +
			"('https://www.youtube.com/embed/KaxRamFgENo')," +
			"('https://www.youtube.com/embed/oi7KrHJki2I')," +
			"('https://www.youtube.com/embed/yDShQA-GLro')," +
			"('https://www.youtube.com/embed/kCgNFxHHgFM')," +
			"('https://www.youtube.com/embed/9GOzEYlRgeQ')," +
			"('https://www.youtube.com/embed/TaxJLWyTiFI')," +
			"('https://www.youtube.com/embed/kHKJ9Mf8UxU')," +
			"('https://www.youtube.com/embed/T-BoDW1_9P4')," +
			"('https://www.youtube.com/embed/k238XpMMn38')," +
			"('https://www.youtube.com/embed/vkUIyOm9hZk')," +
			"('https://www.youtube.com/embed/d9nqtSLOtyU')," +
			"('https://www.youtube.com/embed/wd6Psqzgj5s')," +
			"('https://www.youtube.com/embed/XEVIYNF5Aok')," +
			"('https://www.youtube.com/embed/4t2y_i2pLPY')," +
			"('https://www.youtube.com/embed/o7BkY5lBC6A')," +
			"('https://www.youtube.com/embed/_p06rBGHjMg')," +
			"('https://www.youtube.com/embed/WfVv6eKN4LQ')," +
			"('https://www.youtube.com/embed/wfqaSY6HzgA')," +
			"('https://www.youtube.com/embed/zuFgHtBlm_g')," +
			"('https://www.youtube.com/embed/QDsy8PYN4U4')," +
			"('https://www.youtube.com/embed/qZHm1MVCW8Y')," +
			"('https://www.youtube.com/embed/WePioLV5_cs')," +
			"('https://www.youtube.com/embed/tT-jCqwOZVQ')," +
			"('https://www.youtube.com/embed/pdO9uKpzaYU')," +
			"('https://www.youtube.com/embed/75tgrtvc6rU')," +
			"('https://www.youtube.com/embed/Q-O2wEyz2_c')," +
			"('https://www.youtube.com/embed/9NeknAtFkFg')," +
			"('https://www.youtube.com/embed/sBAFdLDv8O0')," +
			"('https://www.youtube.com/embed/kHm6lTTRdI4')," +
			"('https://www.youtube.com/embed/5x4R6B5i1SU')," +
			"('https://www.youtube.com/embed/V8s6Q5fxpKg')," +
			"('https://www.youtube.com/embed/DwIv7SJCORU')," +
			"('https://www.youtube.com/embed/lkalR1yRNjY')," +
			"('https://www.youtube.com/embed/RJ5goMBD6oc')," +
			"('https://www.youtube.com/embed/Vps90g8bDLQ')," +
			"('https://www.youtube.com/embed/t2Jpe0I5pa4')," +
			"('https://www.youtube.com/embed/KEufcBzoB6Q')," +
			"('https://www.youtube.com/embed/HIKgG6d9YEs')," +
			"('https://www.youtube.com/embed/Zr0nNZwAATM')," +
			"('https://www.youtube.com/embed/ACpP0kNibW0')," +
			"('https://www.youtube.com/embed/cbnIsT-c8dw')," +
			"('https://www.youtube.com/embed/WaXvbkjn-RA')," +
			"('https://www.youtube.com/embed/A67_KSE4d1c')," +
			"('https://www.youtube.com/embed/hcxh0wFB990')," +
			"('https://www.youtube.com/embed/0Dr1iZievjU')," +
			"('https://www.youtube.com/embed/yEpWss9-_uE')," +
			"('https://www.youtube.com/embed/6burVdcR_sc')," +
			"('https://www.youtube.com/embed/2LkTP7kzpl4')," +
			"('https://www.youtube.com/embed/6EiyFbHH7RI')," +
			"('https://www.youtube.com/embed/xfmNv1IeNpM')," +
			"('https://www.youtube.com/embed/YTMBvL1243M')," +
			"('https://www.youtube.com/embed/rO6lQhLpBWQ')," +
			"('https://www.youtube.com/embed/39g-noqtO0A')," +
			"('https://www.youtube.com/embed/co5WgI5zi9s')," +
			"('https://www.youtube.com/embed/6BcH7h3ZDlE')," +
			"('https://www.youtube.com/embed/b8RLWRswQO4')," +
			"('https://www.youtube.com/embed/q1a4RCaKcoo')," +
			"('https://www.youtube.com/embed/GryCbhB69zQ')," +
			"('https://www.youtube.com/embed/Ua8wiPLdEqc')," +
			"('https://www.youtube.com/embed/aMEnAmoXZBE')," +
			"('https://www.youtube.com/embed/0m2T2lEnf3g')," +
			"('https://www.youtube.com/embed/LsILncza29g')," +
			"('https://www.youtube.com/embed/2FZWNJHojCs')," +
			"('https://www.youtube.com/embed/se1mTB6drVs')," +
			"('https://www.youtube.com/embed/N5t1_G2xlIg')," +
			"('https://www.youtube.com/embed/1Gqh37Cf6qM')," +
			"('https://www.youtube.com/embed/51dxbQYAOCU')," +
			"('https://www.youtube.com/embed/7lmk7-H9_Kc')," +
			"('https://www.youtube.com/embed/XWvBIhco_hU')," +
			"('https://www.youtube.com/embed/SRoMR2uLXDI')," +
			"('https://www.youtube.com/embed/cPEukksr5X8')," +
			"('https://www.youtube.com/embed/l5BHsNQlcPs')," +
			"('https://www.youtube.com/embed/l6ACpMUhhjk')," +
			"('https://www.youtube.com/embed/gFcwxTgtagk')," +
			"('https://www.youtube.com/embed/XBWWtEA9c1w')," +
			"('https://www.youtube.com/embed/doujyPMBpYg')," +
			"('https://www.youtube.com/embed/57CYKPBvAtc')",
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
			rollbackErr = tran.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("Failed to perform transaction (%w) and failed to reverse transaction (%v)", err, rollbackErr) 
			}
			return err
		}
	}

	// Commit transaction
	return tran.Commit()
}

func (db *Database) AddUser(user User) error {
	_, err := db.Exec("INSERT INTO users(id, ip) VALUES (?, ?)", user.id[:], user.ip)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetUser(remoteAddr string) (user User, err error) {
	user.ip = remoteAddr

	// Get user from database
	err = db.QueryRow("SELECT id FROM users WHERE ip=? ", user.ip).Scan(&user.id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not found, create new user
			user.id = uuid.New()
			err = db.AddUser(user)
		} else {
			return
		}
	}

	return
}

// GetNextVoteForUser gets the next vote for a user
// If a vote already exists, it will be deleted.
// If there are < 2 options, `vote` will be nil
func (db *Database) GetNextVoteForUser(user User) (vote *VoteOptions, err error) {
	a, b, err := db.findNextPair(user)
	if err != nil {
		// Return nil vote, we don't have enough
		// voting options for this user
		return
	}

	// Exec is 10x-100x slower for some reason.
	// Query has issues committing inserts
	// Locking issue?

	// Should have fixed it via modernc sqlite driver
	// - Arzumify
	vote = &VoteOptions{uuid.New(), a, b}
	_, err = db.Exec(
		"INSERT OR REPLACE INTO active_votes VALUES (?, ?, ?, ?)",
		user.id,
		vote.ID[:],
		a,
		b,
	)
	return
}

// Get new vote options for the user
// Empty a or b strings means not enough available voting options
func (db *Database) findNextPair(user User) (a string, b string, err error) {
	row, err := db.Query(
		"SELECT url FROM videos WHERE url NOT IN (SELECT video_url FROM votes WHERE user_id = ?) ORDER BY random() LIMIT 2",
		user.id,
	)
	if err != nil {
		return
	}
	defer func() {
		err := row.Close()
		if err != nil {
			fmt.Println("Failed to close row", err)
		}
	}()

	if !row.Next() {
		// 0 videos available
		err = fmt.Errorf("no videos available")
		return
	}

	err = row.Scan(&a)
	if err != nil || !row.Next() {
		// 1 video available
		err = fmt.Errorf("only one video available")
		return
	}

	err = row.Scan(&b)
	return
}

func (db *Database) GetCurrentVotingOptionsForUser(user User) (vote *VoteOptions, err error) {
	row, err := db.Query(
		"SELECT a, b FROM active_votes WHERE user_id = ?",
		user.id,
	)
	if err != nil {
		return
	}
	defer func() {
		err := row.Close()
		if err != nil {
			fmt.Println("Failed to close row", err)
		}
	}()

	if !row.Next() {
		// User has no vote options, returning nil
		return
	}

	vote = &VoteOptions{ID: uuid.New()}
	err = row.Scan(
		&vote.A,
		&vote.B,
	)

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
