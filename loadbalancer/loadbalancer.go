package loadbalancer

import (
	"fmt"
	"net"
	"time"
)

const (
	DefaultTimeout = 3

)

type Lb struct {
    Endpoints []string
    Timeout uint8      // health monitor timeout, Default: 3s
    Delay uint8        // health monitor delay, Default:3s
    index int
    pool *pool
    isClose *bool
}

func (cls *Lb) Init() {
	isClose := false
	cls.isClose = &isClose

	p := pool{endpoints: cls.Endpoints}
	p.init()
	cls.pool = &p

	if cls.Delay == 0 {
		cls.Delay = DefaultTimeout
	}
	if cls.Timeout == 0 {
		cls.Timeout = DefaultTimeout
	}
	hM := hm{
		timeout: cls.Timeout,
		delay: cls.Delay,
		endpoints: cls.Endpoints,
		pool: &p,
	}

	go hM.init(cls.isClose)
}

func (cls *Lb) Close() {
	*cls.isClose = true
}

func (cls *Lb) GetEndpoint() (string, error) {
	if *cls.isClose {
		return "", fmt.Errorf("lb is closed")
	}

	endpoints := cls.pool.getAliveEndpoints()

	if len(endpoints) == 1 {
		return endpoints[0], nil
	}

	for idx, ep := range cls.Endpoints {
		if idx > cls.index {
            for _, eP := range endpoints {
            	if ep == eP {
            		cls.index = idx
            		return ep, nil
				}
			}
		}
	}

	for idx, ep := range cls.Endpoints {
		if idx < cls.index {
			for _, eP := range endpoints {
				if ep == eP {
					cls.index = idx
					return ep, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no avaliable endpoint")
}


type pool struct {
	endpoints []string
	members []*member
}

func (cls *pool) init() {
	for _, endpoint := range cls.endpoints {
		cls.members = append(cls.members, &member{endpoint: endpoint, isAlive: endpointIsAlive(endpoint, DefaultTimeout)})
	}
}

func (cls *pool) getAliveEndpoints() []string {
	eps := make([]string, 0)
	for _, me := range cls.members {
		if me.isAlive {
			eps = append(eps, me.endpoint)
		}
	}

	return eps
}

func (cls *pool) setEndpointStatus(endpoint string, status bool) {
	for _, me := range cls.members {
		if me.endpoint == endpoint {
			me.isAlive = status
		}
	}
}


type member struct {
    endpoint string
    isAlive bool
}

type hm struct {
	timeout, delay uint8
	endpoints []string
	pool *pool
}

func (cls *hm) init(isClose *bool) {
	for {
		for _, ep := range cls.endpoints {
			cls.pool.setEndpointStatus(ep, endpointIsAlive(ep, cls.timeout))
		}

		time.Sleep(time.Second * time.Duration(cls.delay))
		if *isClose {
			break
		}
	}
}

func endpointIsAlive(endpoint string, timeout uint8) bool {
	conn, err := net.DialTimeout("tcp", endpoint, time.Second * time.Duration(timeout))
	if err != nil {
		return false
	}

	err = conn.Close()
	return true
}
