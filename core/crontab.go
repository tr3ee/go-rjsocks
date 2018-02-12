package rjsocks

import (
	"sync"
	"time"
)

type CronItem struct {
	Func2run   func()
	LastAccess time.Time
	Interval   time.Duration
}

func NewCronItem(f func(), interval time.Duration) *CronItem {
	return &CronItem{Func2run: f, Interval: interval}
}

type Crontab struct {
	m        sync.Map
	isClosed bool
}

func NewCrontab() *Crontab {
	return new(Crontab)
}

func (c *Crontab) Register(name string, item *CronItem) {
	if !c.Exist(name) {
		c.ForceRegister(name, item)
	}
}

func (c *Crontab) ForceRegister(name string, item *CronItem) {
	item.LastAccess = time.Now()
	c.m.Store(name, item)
}

func (c *Crontab) Exist(name string) bool {
	_, ok := c.m.Load(name)
	return ok
}

func (c *Crontab) Delete(name string) bool {
	if c.Exist(name) {
		c.m.Delete(name)
		return true
	}
	return false
}

func (c *Crontab) UpdateLastAccess(name string, tm time.Time) bool {
	if val, ok := c.m.Load(name); ok {
		item := val.(*CronItem)
		item.LastAccess = tm
		return true
	}
	return false
}

func (c *Crontab) Run() error {
	for t := range time.Tick(1 * time.Second) {
		if c.isClosed {
			break
		}
		rangeFunc := func(n, i interface{}) bool {
			name := n.(string)
			item := i.(*CronItem)
			if t.Sub(item.LastAccess) >= item.Interval {
				c.UpdateLastAccess(name, time.Now())
				go item.Func2run()
			}
			return true
		}
		c.m.Range(rangeFunc)
	}
	return nil
}

func (c *Crontab) Close() {
	c.m.Range(func(key, val interface{}) bool {
		c.m.Delete(key)
		return true
	})
	c.isClosed = true
}
