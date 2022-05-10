package chatbox

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type GUIFrontend struct {
	logger      log.Logger
	conn        Connection
	layoutReady chan struct{}
	gui         *gocui.Gui
}

func (f *GUIFrontend) Start() error {
	logger := f.logger
	g := f.gui

	g.Mouse = true

	f.layoutReady = make(chan struct{})
	g.SetManagerFunc(f.newManagerFunc(func() {
		close(f.layoutReady)
	}))

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, f.quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, f.activateView); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, f.nextView); err != nil {
		logger.Error(err.Error())
		return err
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, f.sendMessageFromInput); err != nil {
		logger.Error(err.Error())
		return err
	}

	stop := make(chan struct{})
	defer close(stop)

	var err error
	select {
	case err = <-f.guiPump():
	case err = <-f.pumpEvents(stop):
	}
	return err
}

func (f *GUIFrontend) Close() {
	f.gui.Close()
}

func (f *GUIFrontend) guiPump() <-chan error {
	g := f.gui
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

func (f *GUIFrontend) pumpEvents(stop <-chan struct{}) <-chan error {
	logger := f.logger
	done := make(chan error)
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case e, ok := <-f.conn.ReceiveEvent():
				if !ok || e == nil {
					err := errors.New("connection closed event channel")
					logger.Error(err.Error())
					done <- err
					return
				}
				switch t := e.(type) {
				case EventUserListUpdate:
					if err := f.setUsers(t.Users); err != nil {
						done <- err
					}

				case EventNewUser:
					msg := fmt.Sprintf(
						"[%s] <<user \"%s\" entered the room>>",
						t.Time.Local(),
						t.Name,
					)
					if err := f.addMessageLine(msg); err != nil {
						done <- err
					}

				case EventNewMessage:
					msg := fmt.Sprintf(
						"[%s %s] >> %s",
						t.Time.Local(),
						t.Sender,
						t.Message,
					)
					if err := f.addMessageLine(msg); err != nil {
						done <- err
					}

				default:
					logger.Warn("unhandled event type",
						map[string]any{
							"type": reflect.TypeOf(e).String(),
						})
				}
			}
		}
	}()
	return done

}

func (f *GUIFrontend) setUsers(usernames []string) error {
	<-f.layoutReady
	g := f.gui
	v, err := g.View("users")
	if err != nil {
		return err
	}
	g.Update(func(g *gocui.Gui) error {
		v.Clear()
		for _, u := range usernames {
			fmt.Fprintln(v, u)
		}
		return nil
	})
	return nil
}

func (f *GUIFrontend) addMessageLine(line string) error {
	<-f.layoutReady
	g := f.gui
	v, err := g.View("messages")
	if err != nil {
		return err
	}
	g.Update(func(g *gocui.Gui) error {
		_, err := fmt.Fprintln(v, line)
		return err
	})
	return nil
}

func (f *GUIFrontend) newManagerFunc(onReady func()) gocui.ManagerFunc {
	once := sync.Once{}
	return func(g *gocui.Gui) error {
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
			if err := f.activateView(g, v); err != nil {
				return err
			}
		}

		once.Do(onReady)
		return nil
	}
}

func (f *GUIFrontend) activateView(g *gocui.Gui, v *gocui.View) error {
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

func (f *GUIFrontend) nextView(g *gocui.Gui, v *gocui.View) error {
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
	return f.activateView(g, nextView)
}

func (f *GUIFrontend) sendMessageFromInput(g *gocui.Gui, v *gocui.View) error {
	v, err := g.View("input")
	bytes, err := ioutil.ReadAll(v)
	if err != nil {
		return err
	}
	f.conn.SendEvent(EventSendMessage{
		EventMeta: *NewEventMetaNow(),
		Message:   string(bytes),
	})
	v.Clear()
	return err
}

func (f *GUIFrontend) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func NewGUIFrontend(conn Connection, logger log.Logger) (*GUIFrontend, error) {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, err
	}
	fe := &GUIFrontend{
		logger: logger,
		conn:   conn,
		gui:    g,
	}
	return fe, nil
}
