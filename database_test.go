package karmabot

import (
	"sort"

	"github.com/kamaln7/karmabot/database"
)

type TestDatabase struct {
	records []database.Points
}

func (t *TestDatabase) InsertPoints(points *database.Points) error {
	t.records = append(t.records, *points)
	return nil
}

func (t *TestDatabase) GetUser(name string) (*database.User, error) {
	foundUser := false
	pointCount := 0
	for _, r := range t.records {
		if r.To == name {
			foundUser = true
			pointCount += r.Points
		}
	}
	if !foundUser {
		return nil, database.ErrNoSuchUser
	}
	return &database.User{
		Name:   name,
		Points: pointCount,
	}, nil
}

func (t *TestDatabase) GetLeaderboard(limit int) (database.Leaderboard, error) {
	us := make(map[string]*database.User)

	for _, r := range t.records {
		u := us[r.To]
		if u == nil {
			u = &database.User{}
		}
		u.Points += r.Points
		us[r.To] = u
	}

	lb := make(database.Leaderboard, 0, len(us))
	for _, u := range us {
		lb = append(lb, u)
	}
	sort.SliceStable(lb, func(i, j int) bool {
		ui := lb[i]
		uj := lb[j]
		if ui.Points == uj.Points && ui.Name < uj.Name {
			return true
		}
		if ui.Points < uj.Points {
			return true
		}
		return false
	})
	return lb[:limit], nil
}

func (t *TestDatabase) GetTotalPoints() (int, error) {
	totalPoints := 0
	for _, r := range t.records {
		p := r.Points
		if p < 0 {
			p = -p
		}
		totalPoints += p
	}
	return totalPoints, nil
}
