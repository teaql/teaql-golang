package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/runtime"
)

type unavailableBusinessClock struct{ reads int }

func (c *unavailableBusinessClock) Now() time.Time {
	c.reads++
	panic("synthetic unavailable business time")
}

func verifyClockDomainSeparation() {
	codec, err := runtime.NewAEADEntityReferenceCodec(1, map[uint32][]byte{1: bytes.Repeat([]byte{0x57}, 32)})
	if err != nil {
		panic(err)
	}
	future := time.Date(2132, 2, 29, 10, 15, 30, 0, time.UTC)
	ctx := runtime.NewUserContext().WithBusinessClock(runtime.NewFixedBusinessClock(future)).WithEntityReferenceCodec(codec)
	require(ctx.BusinessTime().Equal(future), "future business clock was not installed")
	before := time.Now()
	token, err := ctx.EncodeEntityReference("OrderItem", 42, 7, "edit-order", time.Minute)
	if err != nil {
		panic(err)
	}
	claims, err := ctx.DecodeEntityReference(token, "OrderItem", "edit-order")
	require(err == nil, "business future expired a newly issued reference")
	require(claims.IssuedAt.Unix() >= before.Unix() && claims.IssuedAt.Unix() <= time.Now().Unix(), "business clock set security issue time")
	require(claims.ExpiresAt.Sub(claims.IssuedAt) == time.Minute, "business clock changed security lifetime")
	broken := &unavailableBusinessClock{}
	ctx.SetBusinessClock(broken)
	calibrated := false
	func() {
		defer func() { calibrated = recover() == "synthetic unavailable business time" }()
		_ = ctx.BusinessTime()
	}()
	require(calibrated && broken.reads == 1, "unreadable business clock was not calibrated")
	_, err = ctx.DecodeEntityReference(token, "OrderItem", "edit-order")
	require(err == nil && broken.reads == 1, "security verification used unreadable business time")
	_, err = ctx.EncodeEntityReference("OrderItem", 42, 7, "edit-order", time.Minute)
	require(err == nil && broken.reads == 1, "security issue used unreadable business time")
	fmt.Println(`GO_SECURITY_CLOCK_EXAMPLE={"passed":true,"cases":3,"systemIssueTime":true,"validWithFutureBusinessTime":true,"unreadableBusinessCalibrated":true,"securityBusinessClockReads":0}`)
}
