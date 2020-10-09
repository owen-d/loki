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
	return runewidth.FillRight(msg, ln-runewidth.StringWidth(msg))
}

func LPad(msg string, ln int) string {
	return runewidth.FillLeft(msg, ln-runewidth.StringWidth(msg))
}

// ExaExactWidth truncate a message to a particular length or add right padding until the length is hit.
func ExactWidth(msg string, ln int) string {
	return RPad(runewidth.Truncate(msg, ln, " "), ln)
}
