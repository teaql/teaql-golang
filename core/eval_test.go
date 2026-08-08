package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type mockCompany struct {
	name      *string
	loadState *LoadState
}

func (c *mockCompany) evalName() *EvalResult[string] {
	if !c.loadState.IsLoaded("name") {
		return EvalNotLoaded[string]("name", "name")
	}
	if c.name != nil {
		return EvalValue(*c.name)
	}
	return EvalNull[string]()
}

type mockPlatform struct {
	company   *mockCompany
	loadState *LoadState
}

func (p *mockPlatform) evalCompany() *EvalResult[*mockCompany] {
	if !p.loadState.IsLoaded("company") {
		return EvalNotLoaded[*mockCompany]("company", "company")
	}
	if p.company != nil {
		return EvalValue(p.company)
	}
	return EvalNull[*mockCompany]()
}

type mockUser struct {
	platform  *mockPlatform
	loadState *LoadState
}

func (u *mockUser) evalPlatform() *EvalResult[*mockPlatform] {
	if !u.loadState.IsLoaded("platform") {
		return EvalNotLoaded[*mockPlatform]("platform", "platform")
	}
	if u.platform != nil {
		return EvalValue(u.platform)
	}
	return EvalNull[*mockPlatform]()
}

func TestEvalTrackingChainPerfectPath(t *testing.T) {
	company := &mockCompany{
		name:      nil,
		loadState: NewLoadStateNotLoaded(),
	}

	platform := &mockPlatform{
		company:   company,
		loadState: NewLoadStateFullyLoaded(),
	}

	user := &mockUser{
		platform:  platform,
		loadState: NewLoadStateFullyLoaded(),
	}

	result := EvalAndThen(user.evalPlatform(), "platform", func(p *mockPlatform) *EvalResult[string] {
		return EvalAndThen(p.evalCompany(), "company", func(c *mockCompany) *EvalResult[string] {
			return c.evalName()
		})
	})

	assert.Equal(t, EvalResultNotLoaded, result.Type)
	assert.Equal(t, "platform.company.name", result.AttemptedPath)
}

func TestEvalTrackingChainMiddleBreak(t *testing.T) {
	platform := &mockPlatform{
		company:   nil,
		loadState: NewLoadStateNotLoaded(),
	}

	user := &mockUser{
		platform:  platform,
		loadState: NewLoadStateFullyLoaded(),
	}

	result := EvalAndThen(user.evalPlatform(), "platform", func(p *mockPlatform) *EvalResult[string] {
		return EvalAndThen(p.evalCompany(), "company", func(c *mockCompany) *EvalResult[string] {
			return c.evalName()
		})
	})

	assert.Equal(t, EvalResultNotLoaded, result.Type)
	assert.Equal(t, "platform.company", result.AttemptedPath)
}

func TestEvalTrackingChainNormalNull(t *testing.T) {
	company := &mockCompany{
		name:      nil,
		loadState: NewLoadStateFullyLoaded(),
	}

	platform := &mockPlatform{
		company:   company,
		loadState: NewLoadStateFullyLoaded(),
	}

	user := &mockUser{
		platform:  platform,
		loadState: NewLoadStateFullyLoaded(),
	}

	result := EvalAndThen(user.evalPlatform(), "platform", func(p *mockPlatform) *EvalResult[string] {
		return EvalAndThen(p.evalCompany(), "company", func(c *mockCompany) *EvalResult[string] {
			return c.evalName()
		})
	})

	assert.Equal(t, EvalResultNull, result.Type)
}
