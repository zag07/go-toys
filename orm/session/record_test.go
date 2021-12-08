package session

import "testing"

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)

func TestSession_Insert(t *testing.T) {
	s := NewSession().Model(&User{})
	rows, err := s.Insert(user1, user2)
	if rows != 2 || err != nil {
		t.Fatal("failed init test records")
	}
}

func TestSession_Find(t *testing.T) {
	var users []User
	s := NewSession().Model(&User{})
	err := s.Find(&users)
	if err != nil || len(users) != 2 {
		t.Fatal("failed find all records")
	}
}

func TestSession_First(t *testing.T) {
	var user User
	s := NewSession().Model(&User{})
	err := s.First(&user)
	if err != nil || user.Name != "Tom" {
		t.Fatal("failed find first record")
	}
}
