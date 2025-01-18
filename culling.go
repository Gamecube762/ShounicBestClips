package main

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

const minCullVotes = 20
const minCullRatio = .50
const intervalCullTask = 1 * time.Hour

type VoteState = map[string]*VideoStats

type VideoStats struct {
	totalVotes int
	totalScore int
}

func (cv *VideoStats) ShouldCull() bool {
	return cv.totalVotes > minCullVotes &&
		float32(cv.totalScore)/float32(cv.totalVotes) < minCullRatio
}

func taskCullVideos() {
	var err error
	for {
		fmt.Println("Running cull task")
		start := time.Now()

		if err = cullVideos(database); err != nil {
			fmt.Println("Failed to cull videos:" + err.Error())
			sentry.CaptureException(err)
		}

		fmt.Printf("Cull finish at %dms\n", time.Since(start).Milliseconds())
		UpdateUnculledClipTotal()
		fmt.Printf("Counting finish at %dms\n", time.Since(start).Milliseconds())

		time.Sleep(intervalCullTask)
	}
}

func cullVideos(database *Database) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}

	videos := make(VoteState)

	// Count the vote scores and vote counts
	rows, err := tx.Query("SELECT video_url, score FROM votes")
	if err != nil {
		return err
	}

	// IsSingleThreaded := runtime.GOMAXPROCS(0) == 1

	for rows.Next() {
		var url string
		var score int
		err = rows.Scan(&url, &score)
		if err != nil {
			return err
		}

		if videos[url] == nil {
			videos[url] = &VideoStats{}
		}

		videos[url].totalScore += score
		videos[url].totalVotes += 1

		//if IsSingleThreaded {
		//	// Release the loop to allow
		//	// Other goroutines to run.
		//	time.Sleep(250 * time.Nanosecond)
		//}
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	// Reset table to allow videos to return to the queue
	// if we adjust the numbers in prod to be more lenient.
	_, err = tx.Exec("DELETE FROM culled_videos")
	if err != nil {
		tx.Rollback()
		return err
	}

	fmt.Println("Culling Debug:\n\tUrl Votes Score")
	for url, stats := range videos {
		fmt.Printf("\t%s %v ", url, stats)
		fmt.Printf(
			"(%d/%d) = %f > %f\n",
			stats.totalScore,
			stats.totalVotes,
			float32(stats.totalScore)/float32(stats.totalVotes),
			minCullRatio,
		)
		if stats.ShouldCull() {
			_, err = tx.Exec(
				"INSERT OR IGNORE INTO culled_videos VALUES (?)",
				url,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}
