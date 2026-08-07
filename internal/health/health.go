package health

import (
	"context"
	"net"
	"sync"
	"time"
)

const probeTimeout = 500 * time.Millisecond

type Dependency interface {
	Name() string
	Check(context.Context) error
}

type Status struct {
	OK           bool            `json:"ok"`
	Dependencies map[string]bool `json:"dependencies"`
}

type Checker struct {
	dependencies []Dependency
}

func NewChecker(dependencies ...Dependency) *Checker {
	return &Checker{dependencies: dependencies}
}

func (c *Checker) Readiness(ctx context.Context) Status {
	status := Status{OK: true, Dependencies: make(map[string]bool, len(c.dependencies))}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dependency := range c.dependencies {
		wg.Add(1)
		go func(dep Dependency) {
			defer wg.Done()
			probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
			defer cancel()
			ok := dep.Check(probeCtx) == nil

			mu.Lock()
			status.Dependencies[dep.Name()] = ok
			if !ok {
				status.OK = false
			}
			mu.Unlock()
		}(dependency)
	}

	wg.Wait()
	return status
}

type TCPDependency struct {
	name    string
	address string
}

func NewTCPDependency(name, address string) TCPDependency {
	return TCPDependency{name: name, address: address}
}

func (d TCPDependency) Name() string {
	return d.name
}

func (d TCPDependency) Check(ctx context.Context) error {
	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", d.address)
	if err != nil {
		return err
	}
	return connection.Close()
}
