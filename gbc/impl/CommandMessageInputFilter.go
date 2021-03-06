// MIT License
//
// Copyright (c) 2018 dualface
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package impl

import (
    "github.com/dualface/go-gbc/gbc"
)

type (
    CommandMessageInputFilter struct {
        MessageChan chan gbc.RawMessage

        headerBuf    []byte
        headerOffset int
        msg          *CommandMessage
    }
)

func NewCommandMessageInputFilter() *CommandMessageInputFilter {
    p := &CommandMessageInputFilter{
        headerBuf: make([]byte, CommandMessageHeaderLen),
    }
    return p
}

// interface InputFilter

func (f *CommandMessageInputFilter) WriteBytes(input []byte) (output []byte, err error) {
    avail := len(input)

    for ; avail > 0; {

        if f.msg == nil {
            if f.headerOffset < CommandMessageHeaderLen {
                // fill header
                writeLen := avail
                if writeLen > CommandMessageHeaderLen-f.headerOffset {
                    writeLen = CommandMessageHeaderLen - f.headerOffset
                }

                copy(f.headerBuf[f.headerOffset:], input[0:writeLen])
                f.headerOffset += writeLen
                if f.headerOffset < CommandMessageHeaderLen {
                    // if header not filled, return
                    return
                }

                input = input[writeLen:]
                avail -= writeLen
            }

            // header is ready, create msg
            f.headerOffset = 0
            f.msg, err = NewCommandMessageFromHeaderBuf(f.headerBuf)
            if err != nil {
                return
            }
        }

        // determine write len
        appendLen := f.msg.RemainsBytes()
        if appendLen > avail {
            appendLen = avail
        }
        avail -= appendLen

        // write to message
        _, err = f.msg.WriteBytes(input[0:appendLen])
        if err != nil {
            return
        }
        input = input[appendLen:]

        // if get a message, send it
        if f.msg.RemainsBytes() == 0 {
            if f.MessageChan != nil {
                f.MessageChan <- f.msg
            }
            f.msg = nil
        }
    }

    return
}

func (f *CommandMessageInputFilter) SetRawMessageChannel(mc chan gbc.RawMessage) {
    f.MessageChan = mc
}
