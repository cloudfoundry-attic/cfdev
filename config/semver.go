package config

import (
	"strconv"
	"strings"
)

type Version struct {
	Major    int
	Minor    int
	Build    int
	Original string
}

func NewSemver(v string) (*Version, error) {
	s := &Version{Original: v}
	arr := strings.SplitN(v, ".", 3)
	var err error
	if len(arr) >= 1 && v != "" {
		s.Major, err = strconv.Atoi(arr[0])
		if err != nil {
			return nil, err
		}
	}
	if len(arr) >= 2 {
		s.Minor, err = strconv.Atoi(arr[1])
		if err != nil {
			return nil, err
		}
	}
	if len(arr) >= 3 {
		parts := strings.SplitN(arr[2], "-", 2)
		s.Build, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}
