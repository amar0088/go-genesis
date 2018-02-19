// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"net/url"
	"testing"
	"time"
	//	"github.com/GenesisKernel/go-genesis/packages/converter"
)

func TestTimeLimit(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	all := 2
	sendList := func() {
		for i := 0; i < all; i++ {
			if err := postTx(`TestPars`, &url.Values{`Node`: {`1` /*converter.IntToStr(i)*/}, `nowait`: {`true`}}); err != nil {
				t.Error(err)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	sendList()
	node = 1
	all = 5
	sendList()
	node = 2
	all = 9
	sendList()

}
