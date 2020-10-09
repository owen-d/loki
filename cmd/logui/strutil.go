package main

import "github.com/mattn/go-runewidth"

func intersperse(xs, ys []string) (res []string) {
	for i := 0; i < len(xs) && i < len(ys); i++ {
		res = append(res, xs[i])
		res = append(res, ys[i])
	}
	return res
}
func CenterTo(msg string, ln int) string {
	msgLn := runewidth.StringWidth(msg)
	rem := ln - msgLn
	if rem < 1 {
		return msg
	}

	div := rem / 2
	msg = runewidth.FillLeft(msg, msgLn+div)
	msg = runewidth.FillRight(msg, ln)
	return msg
}

func RPad(msg string, ln int) string {
	return runewidth.FillRight(msg, ln)
}

func LPad(msg string, ln int) string {
	return runewidth.FillLeft(msg, ln)
}

// Truncate removes any overflow past a desired length. It's possible for the result
// to be shorter than the desired length.
func Truncate(msg string, ln int) string {
	if runewidth.StringWidth(msg) <= ln {
		return msg
	}

	r := []rune(msg)
	var i, rw int
	for ; i < len(r); i++ {
		rw += runewidth.RuneWidth(r[i])
		if rw > ln {
			break
		}
	}
	return string(r[0:i])
}

// ExaExactWidth truncate a message to a particular length or add right padding until the length is hit.
func ExactWidth(msg string, ln int) string {
	return RPad(Truncate(msg, ln), ln)
}
