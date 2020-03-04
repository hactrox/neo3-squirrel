package time

import "testing"

func Test(t *testing.T) {
	testCases := map[uint64]string{
		0:    "",
		1:    "01s",
		59:   "59s",
		60:   "01m 00s",
		61:   "01m 01s",
		3599: "59m 59s",
		3600: "01h 00m 00s",
		3601: "01h 00m 01s",
		3659: "01h 00m 59s",
		3660: "01h 01m 00s",
		3661: "01h 01m 01s",
	}

	for duration, want := range testCases {
		get := ParseSeconds(duration)

		if get != want {
			t.Fatalf("Incorrect result from `Parse(%d)`, get=%s, want=%s",
				duration, get, want)
		}
	}
}
