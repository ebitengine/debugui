// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package debugui

func IDPartFromCaller() string {
	pc := caller()
	return idPartFromCaller(pc)
}

func (d *DebugUI) ContainerCounter() int {
	return len(d.ctx.idToContainer)
}

type TextFieldState struct {
	s *textFieldState
}

func NewTextFieldState(text string) TextFieldState {
	s := newTextFieldState()
	s.setText(text)
	return TextFieldState{s: s}
}

func (t TextFieldState) Text() string {
	return t.s.text()
}

func (t TextFieldState) TextForRendering() string {
	return t.s.textForRendering()
}

func (t TextFieldState) Selection() (start, end int) {
	return t.s.selectionStart(), t.s.selectionEnd()
}

func (t TextFieldState) MoveCaretTo(pos int, extend bool) {
	t.s.moveCaretTo(pos, extend)
}

func (t TextFieldState) MoveCaretLeft(extend bool) {
	t.s.moveCaretLeft(extend)
}

func (t TextFieldState) MoveCaretRight(extend bool) {
	t.s.moveCaretRight(extend)
}

func (t TextFieldState) MoveCaretWordLeft(extend bool) {
	t.s.moveCaretWordLeft(extend)
}

func (t TextFieldState) MoveCaretWordRight(extend bool) {
	t.s.moveCaretWordRight(extend)
}

func (t TextFieldState) SelectAll() {
	t.s.selectAll()
}

func (t TextFieldState) SelectWordAt(pos int) {
	t.s.selectWordAt(pos)
}

func (t TextFieldState) HandleClick(pos int, extend bool, now, interval int64) {
	t.s.handleClick(pos, extend, now, interval)
}

func (t TextFieldState) Dragging() bool {
	return t.s.dragging
}

func (t TextFieldState) DeleteBackward() {
	t.s.deleteBackward()
}

func (t TextFieldState) DeleteForward() {
	t.s.deleteForward()
}

func WordRangeAt(s string, pos int) (start, end int) {
	return wordRangeAt(s, pos)
}

func TextIndexFromX(str string, x int) int {
	return textIndexFromX(str, x)
}

func TextWidth(str string) int {
	return textWidth(str)
}
