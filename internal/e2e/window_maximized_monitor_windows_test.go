//go:build windows && (amd64 || arm64)

package e2e

import (
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// TestMaximizedWindowKeepsItsMonitorAfterRestore minimizes and restores a
// maximized window with a hidden title bar that is maximized on a display
// which is not the primary one: the client area must come back on the
// display the window is maximized on.
//
// A minimized window rests off screen, where the display its window
// resolves to is not the one it is restored to: the two differ only with a
// second display, so a machine with one display (CI has one) skips this.
func TestMaximizedWindowKeepsItsMonitorAfterRestore(t *testing.T) {
	var second mygo.Display
	for _, d := range mygo.Screen.Displays() {
		if !d.Primary {
			second = d
			break
		}
	}
	if second.Bounds.Width == 0 {
		t.Skip("needs a second display, to tell the display of the window from the one it is restored to")
	}

	w := newWindow(t, mygo.WindowOptions{
		Title:          "Maximized on a second display",
		TitleBarStyle:  mygo.TitleBarHiddenInset,
		TitleBarHeight: 40,
		Content: ui.View(func(c *ui.Context) {
			ui.Text(c, "Maximized on a second display")
		}),
	})
	// Move it onto the second display, and maximize it there.
	work := second.WorkArea
	w.SetBounds(mygo.Rectangle{X: work.X + 40, Y: work.Y + 40, Width: 600, Height: 400})
	w.Maximize()
	eventually(t, "the window maximized on the second display", func() bool {
		return w.IsMaximized() && onDisplay(w.ContentBounds(), second)
	})

	w.Minimize()
	eventually(t, "a minimized window", w.IsMinimized)
	w.Restore()
	eventually(t, "a restored window", func() bool { return !w.IsMinimized() })
	deadline := time.Now().Add(2 * time.Second)
	for {
		b := w.ContentBounds()
		if onDisplay(b, second) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the client area came back at %+v, off the display %+v the window is maximized on", b, second.Bounds)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// onDisplay reports whether r lies within d.
func onDisplay(r mygo.Rectangle, d mygo.Display) bool {
	return r.X >= d.Bounds.X && r.X+r.Width <= d.Bounds.X+d.Bounds.Width &&
		r.Y >= d.Bounds.Y && r.Y+r.Height <= d.Bounds.Y+d.Bounds.Height
}
