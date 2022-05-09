package ui

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type Client interface {
	SendMessage(m message.Message)
	ReceiveMessage() <-chan message.Message
	Stopped() <-chan struct{}
}

type UI struct {
	logger log.Logger
	client Client
	gui    *gocui.Gui
}

func (u *UI) Start() error {
	logger := u.logger
	g := u.gui

	g.Mouse = true
	g.SetManagerFunc(u.layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, u.quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, u.activateView); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, u.nextView); err != nil {
		logger.Error(err.Error())
		return err
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, u.sendMessageFromInput); err != nil {
		logger.Error(err.Error())
		return err
	}

	var err error
	stop := make(chan bool)

	select {
	case err = <-u.messagePump(stop):
	case err = <-u.guiPump():
	}

	close(stop)
	return err
}

func (u *UI) Close() {
	u.gui.Close()
}

func (u *UI) guiPump() <-chan error {
	g := u.gui
	done := make(chan error)
	go func() {
		defer close(done)
		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			done <- err
		} else {
			done <- nil
		}
	}()
	return done
}

func (u *UI) messagePump(stop <-chan bool) <-chan error {
	done := make(chan error)
	g := u.gui
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
			case m := <-u.client.ReceiveMessage():
				g.Update(func(g *gocui.Gui) error {

					switch m.Name {
					case message.USER_LIST:
						v, err := g.View("users")
						if err != nil {
							return err
						}
						data, err := message.GetData[[]string](m)
						if err != nil {
							return err
						}
						contents := strings.Join(data, "\n")
						v.Clear()
						fmt.Fprint(v, contents)

					case message.NEW_USER:
						v, err := g.View("messages")
						if err != nil {
							return err
						}
						data, err := message.GetData[string](m)
						if err != nil {
							return err
						}
						fmt.Fprintf(
							v,
							"[%s] <<user \"%s\" entered the room>>\n",
							time.Now().Local(),
							data,
						)

					case message.NEW_MESSAGE:
						v, err := g.View("messages")
						if err != nil {
							return err
						}
						data, err := message.GetData[message.NewMessageData](m)
						if err != nil {
							return err
						}
						fmt.Fprintf(
							v,
							"[%s %s] >> %s\n",
							data.Time.Local(),
							data.Sender,
							data.Message,
						)
					}
					return nil
				})
			}
		}
	}()
	return done
}

func (u *UI) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	x0 := 5
	x1 := maxX - 5
	y0 := 5
	y1 := maxY - 5

	if v, err := g.SetView("messages", x0, y0, x1-36, y1-4, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return nil
		}
		v.Title = "Messages"
		v.Wrap = true
		v.Autoscroll = true
		v.Frame = true
	}

	if v, err := g.SetView("users", x1-35, y0, x1-1, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return nil
		}
		v.Title = "Users"
		v.Wrap = true
		v.Autoscroll = true
		v.Frame = true
	}

	if v, err := g.SetView("input", x0, y1-3, x1-36, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return nil
		}
		v.Title = "Input"
		v.Wrap = true
		v.Frame = true
		// v.Autoscroll = true
		v.Editable = true
		if err := u.activateView(g, v); err != nil {
			return err
		}
	}

	return nil
}

func (u *UI) activateView(g *gocui.Gui, v *gocui.View) error {
	g.SetCurrentView(v.Name())
	for _, v := range g.Views() {
		if g.CurrentView() == v {
			v.SelFgColor = gocui.ColorRed
			v.SelBgColor = gocui.ColorBlue
		} else {
			v.SelFgColor = gocui.ColorDefault
			v.SelBgColor = gocui.ColorDefault
		}
	}
	if v.Name() == "input" {
		g.Cursor = true
	} else {
		g.Cursor = false
	}
	return nil
}

func (u *UI) nextView(g *gocui.Gui, v *gocui.View) error {
	all := g.Views()
	if len(all) == 0 {
		return nil
	}
	cur := g.CurrentView()
	nextView := all[0]
	for i, v := range all {
		if cur != nil && i+1 < len(all) && cur.Name() == v.Name() {
			nextView = all[i+1]
		}
	}
	return u.activateView(g, nextView)
}

func (u *UI) sendMessageFromInput(g *gocui.Gui, v *gocui.View) error {
	v, err := g.View("input")
	bytes, err := ioutil.ReadAll(v)
	if err != nil {
		return err
	}
	msg, err := message.NewMessage(message.SEND_MESSAGE, string(bytes))
	if err != nil {
		return err
	}
	u.client.SendMessage(msg)
	v.Clear()
	return err
}

func (u *UI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func NewUI(client Client, logger log.Logger) (*UI, error) {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return &UI{}, err
	}
	ui := &UI{logger: logger, client: client, gui: g}
	return ui, nil
}
