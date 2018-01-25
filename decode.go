// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

// parseBool is similar to strconv.ParseBool but supports the additional cases of allowing yes, Yes,
// YES, on, On, ON, off, Off, OFF, no, No, and NO as boolean values. It doesn't support wonky-case,
// such as oFf, tRuE, etc.
func parseBool(val string) (bool, error) {
	switch val {
	case "1", "t", "T", "true", "True", "TRUE", "yes", "Yes", "YES", "on", "On", "ON":
		return true, nil
	case "0", "f", "F", "false", "False", "FALSE", "no", "No", "NO", "off", "Off", "OFF":
		return false, nil
	default:
		return false, mksyntaxerr(val, ErrInvalidBool)
	}
}
