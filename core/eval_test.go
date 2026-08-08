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

func TestLoadState(t *testing.T) {
	lsNotLoaded := NewLoadStateNotLoaded()
	assert.False(t, lsNotLoaded.IsLoaded("test"))

	lsFullyLoaded := NewLoadStateFullyLoaded()
	assert.True(t, lsFullyLoaded.IsLoaded("test"))

	lsPartial := NewLoadStatePartial([]string{"field1", "field2"})
	assert.True(t, lsPartial.IsLoaded("field1"))
	assert.True(t, lsPartial.IsLoaded("field2"))
	assert.False(t, lsPartial.IsLoaded("field3"))
}

func TestEvalAndThenNullAndNotLoaded(t *testing.T) {
	resNull := EvalNull[string]()
	res2 := EvalAndThen(resNull, "f1", func(s string) *EvalResult[int] {
		return EvalValue(1)
	})
	assert.Equal(t, EvalResultNull, res2.Type)

	resNotLoaded := EvalNotLoaded[string]("node", "f1")
	res3 := EvalAndThen(resNotLoaded, "f2", func(s string) *EvalResult[int] {
		return EvalValue(1)
	})
	assert.Equal(t, EvalResultNotLoaded, res3.Type)
	assert.Equal(t, "f1", res3.AttemptedPath)

	// nextRes.AttemptedPath == ""
	res4 := EvalAndThen(EvalValue(""), "field", func(s string) *EvalResult[int] {
		return EvalNotLoaded[int]("node", "")
	})
	assert.Equal(t, EvalResultNotLoaded, res4.Type)
	assert.Equal(t, "field", res4.AttemptedPath)
}

func TestEvalMap(t *testing.T) {
	resValue := EvalValue(10)
	mapped1 := EvalMap(resValue, func(i int) string {
		return "a"
	})
	assert.Equal(t, EvalResultValue, mapped1.Type)
	assert.Equal(t, "a", mapped1.Value)

	resNull := EvalNull[int]()
	mapped2 := EvalMap(resNull, func(i int) string {
		return "a"
	})
	assert.Equal(t, EvalResultNull, mapped2.Type)

	resNotLoaded := EvalNotLoaded[int]("node", "path")
	mapped3 := EvalMap(resNotLoaded, func(i int) string {
		return "a"
	})
	assert.Equal(t, EvalResultNotLoaded, mapped3.Type)
	assert.Equal(t, "path", mapped3.AttemptedPath)
}

func TestIsLoadedInvalid(t *testing.T) {
	ls := &LoadState{Type: LoadStateType(999)}
	assert.False(t, ls.IsLoaded("test"))
}

func TestEvalInvalid(t *testing.T) {
	res := &EvalResult[string]{Type: EvalResultType(999)}
	
	mapped := EvalMap(res, func(s string) string { return s })
	assert.Equal(t, EvalResultNull, mapped.Type)
	
	andThen := EvalAndThen(res, "field", func(s string) *EvalResult[int] {
		return EvalValue(1)
	})
	assert.Equal(t, EvalResultNull, andThen.Type)
}

