package systemd

import (
	"os/exec"
	"strings"
	"time"

	"barista.run/bar"
	"barista.run/base/value"
	"barista.run/colors"
	"barista.run/outputs"
	"barista.run/pango"
	"barista.run/timing"
)

// Module presents a toggle-able module
type Module struct {
	toggleValue value.Value // of string
	toggleFunc  func() string
	clickFunc   func()
	outputFunc  func(string) *bar.Segment
	scheduler   *timing.Scheduler
}

func New(toggleFunc func() string, clickFunc func(), outputFunc func(string) *bar.Segment, every time.Duration) *Module {
	return &Module{
		toggleFunc: toggleFunc,
		clickFunc:  clickFunc,
		outputFunc: outputFunc,
		scheduler:  timing.NewScheduler().Every(every),
	}
}

func (m *Module) Stream(s bar.Sink) {
	m.refresh()
	toggleValue := m.toggleValue.Get().(string)
	toggleValueSub, done := m.toggleValue.Subscribe()
	defer done()
	for {
		s.Output(
			m.outputFunc(toggleValue).OnClick(m.click),
		)
		select {
		case <-toggleValueSub:
			toggleValue = m.toggleValue.Get().(string)
		case <-m.scheduler.C:
			m.refresh()
		}
	}
}

func (m *Module) click(e bar.Event) {
	m.clickFunc()
	m.refresh()
}

func (m *Module) refresh() {
	v := m.toggleFunc()
	m.toggleValue.Set(v)
}

type SystemdModule struct {
	*Module
	serviceName string
}

func NewSystemdUserService(serviceName string) *SystemdModule {
	m := &SystemdModule{
		serviceName: serviceName,
	}
	m.Module = New(
		m.toggleSystemdUserService,
		m.clickSystemdUserService,
		m.outputSystemdUserService,
		time.Second*5,
	)
	return m
}

func (m *SystemdModule) toggleSystemdUserService() string {
	out, _ := exec.Command("systemctl", "--user", "is-active", m.serviceName).Output()
	return strings.TrimSpace(string(out))
}

func (m *SystemdModule) outputSystemdUserService(serviceState string) *bar.Segment {
	var stateColor string
	switch serviceState {
	case "active":
		stateColor = "#238555"
	case "inactive":
		stateColor = "#972822"
	default:
		stateColor = "#f70"
	}
	return outputs.Pango(
		pango.Icon("mdi-arrow-decision").Alpha(0.6),
		pango.Text(m.serviceName).Color(colors.Hex(stateColor)),
	)
}

func (m *SystemdModule) clickSystemdUserService() {
	out, _ := exec.Command("systemctl", "--user", "is-active", m.serviceName).Output()
	var toggleCmd string

	switch strings.TrimSpace(string(out)) {
	case "active":
		toggleCmd = "stop"
	case "inactive":
		toggleCmd = "start"
	default:
		toggleCmd = "restart"
	}
	_, _ = exec.Command("systemctl", "--user", toggleCmd, m.serviceName).CombinedOutput()

	m.refresh()
}
