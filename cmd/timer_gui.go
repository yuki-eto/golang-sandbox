package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/martinlindhe/notify"
	"github.com/zserge/webview"
)

type Timer struct {
	target *time.Time
	Now    string    `json:"now"`
	Remain string    `json:"remain"`
	EndCh  chan bool `json:"-"`
}

func NewTimer(t *time.Time, d time.Duration) *Timer {
	target := t.Add(d)
	return &Timer{
		target: &target,
		EndCh:  make(chan bool, 1),
	}
}

func (t *Timer) Update() {
	now := time.Now()
	t.Now = now.Format("2006/01/02 15:04:05 MST")

	if now.After(*t.target) {
		t.EndCh <- true
		return
	}

	diff := t.target.Sub(now)
	hour := uint64(diff.Hours()) % 24
	min := uint64(diff.Minutes()) % 60
	sec := uint64(diff.Seconds()) % 60
	msec := (uint64(diff.Milliseconds()) % 1000) / 100
	t.Remain = fmt.Sprintf("Remain %02d:%02d:%02d.%01d", hour, min, sec, msec)
}

func main() {
	const html = `
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <style>
      html, body {
        margin: 0;
        height: 100%%;
        overflow: hidden;
        -webkit-user-select: none;
      }
    </style>
    <title>Hello, world!</title>
  </head>
  <body>
    <div id="root">
      <div class="jumbotron text-center text-monospace">
        <h1 class="display-4 text-monospace">{{ timer.remain }}</h1>
        <p class="lead">{{ timer.now }}</p>
      </div>
    </div>

<script src="https://code.jquery.com/jquery-3.4.1.slim.min.js" integrity="sha384-J6qa4849blE2+poT4WnyKhv5vZF5SrPo0iEjwBvKU7imGFAV0wwj1yYfoRSJoZ+n" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/popper.js@1.16.0/dist/umd/popper.min.js" integrity="sha384-Q6E9RHvbIyZFJoft+2mJbHaEWldlvI9IOYy5n3zV9zzTtmI3UksdQRVvoxMfooAo" crossorigin="anonymous"></script>
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/js/bootstrap.min.js" integrity="sha384-wfSDF2E50Y2D1uUdj0O3uMBJnjuUD4Ih7YwaYd1iqfktj0Uod8GCExl3Og8ifwB6" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/vue@2.6.11"></script>
<script>
const vm = new Vue({
  el: '#root',
  data: { timer: Timer.data },
  mounted: function() {
   const self = this;
   setInterval(() => {
     Timer.update();
     self.timer = Timer.data;
   }, 100);
  },
});
</script>
  </body>
</html>
`

	flag.Parse()
	min := flag.Arg(0)
	if min == "" {
		fmt.Println("please input target minutes")
		return
	}

	m, err := strconv.ParseUint(min, 10, 64)
	if err != nil {
		panic(err)
	}
	if m < 1 || m > 1440 {
		fmt.Println("target min have to 1 - 1440")
		return
	}
	now := time.Now()
	d := time.Duration(m) * time.Minute
	timer := NewTimer(&now, d)
	timer.Update()

	w := webview.New(webview.Settings{
		Title:                  "webview",
		URL:                    "data:text/html," + url.PathEscape(html),
		Width:                  640,
		Height:                 240,
		Resizable:              false,
		Debug:                  true,
		ExternalInvokeCallback: nil,
	})
	defer w.Exit()
	w.Dispatch(func() {
		w.Bind("Timer", timer)
	})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

		for {
			select {
			case <-timer.EndCh:
				notify.Alert("Timer", fmt.Sprintf("%d minutes passed", m), "", "")
				w.Exit()
			case <-sig:
				w.Exit()
			}
		}
	}()

	w.Run()
}
