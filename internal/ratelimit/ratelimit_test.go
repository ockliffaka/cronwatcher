package ratelimit_test

import (
	"testing"
	"time"

	"github.com/cronwatcher/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if !l.Allow("backup") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallBlocked(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("backup")
	if l.Allow("backup") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestAllow_DifferentJobsAreIndependent(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("jobA")
	if !l.Allow("jobB") {
		t.Fatal("expected different job to be allowed independently")
	}
}

func TestAllow_PermittedAfterCooldown(t *testing.T) {
	l := ratelimit.New(50 * time.Millisecond)
	l.Allow("backup")
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("backup") {
		t.Fatal("expected call to be allowed after cooldown expires")
	}
}

func TestReset_ClearsRecord(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("backup")
	l.Reset("backup")
	if !l.Allow("backup") {
		t.Fatal("expected call to be allowed after reset")
	}
}

func TestLastSent_UnknownJob(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	_, ok := l.LastSent("unknown")
	if ok {
		t.Fatal("expected no record for unknown job")
	}
}

func TestLastSent_KnownJob(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	before := time.Now()
	l.Allow("backup")
	t2, ok := l.LastSent("backup")
	if !ok {
		t.Fatal("expected record for known job")
	}
	if t2.Before(before) {
		t.Fatalf("recorded time %v is before test start %v", t2, before)
	}
}
