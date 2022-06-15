package db

import (
	"testing"
	"time"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

var (
	followersFixture1 = []models.User{
		{
			ID:       1,
			UserName: "user1",
			FullName: "test user 1",
		},
		{
			ID:       2,
			UserName: "user2",
			FullName: "test user 2",
		},
	}
	followersFixture2 = []models.User{
		{
			ID:       1,
			UserName: "user1",
			FullName: "test user 1",
		},
		{
			ID:       2,
			UserName: "user2",
			FullName: "test user 2",
		},
		{
			ID:       3,
			UserName: "user3",
			FullName: "test user 3",
		},
	}

	followingsFixture1 = []models.User{
		{
			ID:       3,
			UserName: "user3",
			FullName: "test user 3",
		},
		{
			ID:       4,
			UserName: "user4",
			FullName: "test user 4",
		},
	}
)

func resetBatchesTime(s []models.UsersBatch) []models.UsersBatch {
	for i := range s {
		s[i] = resetBatchTime(s[i])
	}

	return s
}

func resetBatchTime(s models.UsersBatch) models.UsersBatch {
	s.CreatedAt = time.Time{}

	return s
}

func setUpLocalDBWithFixtures(t testing.TB) DB {
	t.Helper()

	now := time.Now()

	fixtures := map[models.UsersBatchType][]models.UsersBatch{
		models.UsersBatchTypeFollowers: {
			{
				Users:     followersFixture1,
				Type:      models.UsersBatchTypeFollowers,
				CreatedAt: now.AddDate(0, 0, -2),
			},
			{
				Users:     followersFixture2,
				Type:      models.UsersBatchTypeFollowers,
				CreatedAt: now,
			},
		},
		models.UsersBatchTypeFollowings: {
			{
				Users:     followingsFixture1,
				Type:      models.UsersBatchTypeFollowings,
				CreatedAt: time.Time{},
			},
		},
	}

	db := newLocalDB()
	db.users = fixtures

	return db
}
