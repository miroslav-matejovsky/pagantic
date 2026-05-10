package orchestrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMajorityVoting_Empty(t *testing.T) {
	v := MajorityVoting{}
	_, _, err := v.Vote(nil)
	require.Error(t, err)
}

func TestMajorityVoting_Single(t *testing.T) {
	v := MajorityVoting{}
	winner, conf, err := v.Vote([]string{"hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", winner)
	require.Equal(t, 1.0, conf)
}

func TestMajorityVoting_Unanimous(t *testing.T) {
	v := MajorityVoting{}
	winner, conf, err := v.Vote([]string{"a", "a", "a"})
	require.NoError(t, err)
	require.Equal(t, "a", winner)
	require.Equal(t, 1.0, conf)
}

func TestMajorityVoting_Majority(t *testing.T) {
	v := MajorityVoting{}
	winner, conf, err := v.Vote([]string{"a", "b", "a"})
	require.NoError(t, err)
	require.Equal(t, "a", winner)
	require.InDelta(t, 0.666, conf, 0.01)
}

func TestMajorityVoting_AllDifferent(t *testing.T) {
	v := MajorityVoting{}
	winner, conf, err := v.Vote([]string{"a", "b", "c"})
	require.NoError(t, err)
	require.NotEmpty(t, winner) // one of them wins
	require.InDelta(t, 0.333, conf, 0.01)
}

func TestUnanimityVoting_Empty(t *testing.T) {
	v := UnanimityVoting{}
	_, _, err := v.Vote(nil)
	require.Error(t, err)
}

func TestUnanimityVoting_Single(t *testing.T) {
	v := UnanimityVoting{}
	winner, conf, err := v.Vote([]string{"hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", winner)
	require.Equal(t, 1.0, conf)
}

func TestUnanimityVoting_Unanimous(t *testing.T) {
	v := UnanimityVoting{}
	winner, conf, err := v.Vote([]string{"x", "x", "x"})
	require.NoError(t, err)
	require.Equal(t, "x", winner)
	require.Equal(t, 1.0, conf)
}

func TestUnanimityVoting_Disagreement(t *testing.T) {
	v := UnanimityVoting{}
	_, _, err := v.Vote([]string{"a", "a", "b"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "differs")
}
