package ratelimit

import (
	gc "gopkg.in/check.v1"
	"time"
)

func (rateLimitSuite) TestRate1(c *gc.C) {
	log := newUsageLog(5)
	now := time.Now()
	log.lastSeconds = now.Unix()

	log.usage(now, 5*time.Second, 10)

	c.Assert(log.rightIndex, gc.Equals, 4)
	c.Assert(log.log[0], gc.Equals, int64(2))
	c.Assert(log.log[1], gc.Equals, int64(2))
	c.Assert(log.log[2], gc.Equals, int64(2))
	c.Assert(log.log[3], gc.Equals, int64(2))
	c.Assert(log.log[4], gc.Equals, int64(2))
	now = now.Add(5 * time.Second)

	now = now.Add(2 * time.Second)
	log.usage(now, 0, 3)

	c.Assert(log.rightIndex, gc.Equals, 1)
	c.Assert(log.log[0], gc.Equals, int64(0))
	c.Assert(log.log[1], gc.Equals, int64(3))
	c.Assert(log.log[2], gc.Equals, int64(2))
	c.Assert(log.log[3], gc.Equals, int64(2))
	c.Assert(log.log[4], gc.Equals, int64(2))

	now = now.Add(20 * time.Second)
	log.usage(now, 0, 5)
	c.Assert(log.log[0], gc.Equals, int64(0))
	c.Assert(log.log[1], gc.Equals, int64(5))
	c.Assert(log.log[2], gc.Equals, int64(0))
	c.Assert(log.log[3], gc.Equals, int64(0))
	c.Assert(log.log[4], gc.Equals, int64(0))

}

func (rateLimitSuite) TestRate2(c *gc.C) {
	log := newUsageLog(5)
	now := time.Now()
	log.lastSeconds = now.Unix()

	log.usage(now, 0, 10)

	now = now.Add(1 * time.Second)
	log.usage(now, 0, 20)

	now = now.Add(1 * time.Second)
	log.usage(now, 0, 30)

	now = now.Add(1 * time.Second)
	log.usage(now, 0, 40)

	now = now.Add(1 * time.Second)
	log.usage(now, 0, 50)

	c.Assert(log.usageRate(now), gc.Equals, int64(150/5))
	now = now.Add(1 * time.Second)
	c.Assert(log.usageRate(now), gc.Equals, int64(140/5))
	now = now.Add(1 * time.Second)
	c.Assert(log.usageRate(now), gc.Equals, int64(120/5))
	now = now.Add(1 * time.Second)
	c.Assert(log.usageRate(now), gc.Equals, int64(90/5))
	now = now.Add(1 * time.Second)
	c.Assert(log.usageRate(now), gc.Equals, int64(50/5))
	now = now.Add(1 * time.Second)
	c.Assert(log.usageRate(now), gc.Equals, int64(0/5))

}
