package channel

import (
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func ConnectUser(r *room.Room, u *user.User) error {
	in, out, err := r.Connect(u.Uuid())
	if err != nil {
		return err
	}
	if err := u.Connect(in, out); err != nil {
		return err
	}
	return nil
}
