package version

import (
	"fmt"
	"regexp"
	"strconv"
	"errors"
	"strings"
)

// DefaultPrefix that is recognized and ignored by the parser
const DefaultPrefix = "v"

// Predefined format strings to be used with the Format function
const (
	FullFormat    = "x.y.z-p+m"
	NoMetaFormat  = "x.y.z-p"
	NoPreFormat   = "x.y.z"
	NoPatchFormat = "x.y"
	NoMinorFormat = "x"
	ReleaseCandidate = "x.y.z-r"
)

type buffer []byte

func (b *buffer) AppendInt(i int, sep byte) {
	b.AppendString(strconv.FormatInt(int64(i), 10), sep)
}

func (b *buffer) AppendString(s string, sep byte) {
	if len(s) > 0 && len(*b) > 0 {
		*b = append(*b, sep)
	}
	*b = append(*b, s...)
}

// Version holds the parsed components of git describe
type Version struct {
	Prefix     string
	Major      int
	Minor      int
	Patch      int
	preRelease string
	Commits    int
	Meta       string
	releaseCandidate int
}

// Format returns a string representation of the version including the parts
// defined in the format string. The format can have the following components:
// * x -> major version
// * y -> minor version
// * z -> patch version
// * p -> pre-release
// * m -> metadata
// * r -> release-candidate
// x, y and z are separated by a dot. p is seprated by a hyphen and m by a plus sing.
// E.g.: x.y.z-p+m or x.y
func (v Version) Format(format string) (string, error) {
	re := regexp.MustCompile(
		`(?P<major>x)(?P<minor>\.y)?(?P<patch>\.z)?(?P<pre>-p)?(?P<release_candidate>-r)?(?P<meta>\+m)?`)

	matches := re.FindStringSubmatch(format)
	if matches == nil {
		return "", fmt.Errorf("invalid format: %s", format)
	}

	var buf buffer

	names := re.SubexpNames()
	for i := 0; i < len(matches); i++ {
		if len(matches[i]) == 0 {
			continue
		}
		switch names[i] {
		case "major":
			buf.AppendInt(v.Major, '.')
		case "minor":
			buf.AppendInt(v.Minor, '.')
		case "patch":
			patch := v.Patch
			if v.Commits > 0 && v.preRelease == "" {
				patch++
			}
			buf.AppendInt(patch, '.')
		case "pre":
			buf.AppendString(v.PreRelease(), '-')
		case "release_candidate":
			releaseCandidate, err := v.ReleaseCandidate()
			if err != nil {
				return "", err
			}
			buf.AppendString(releaseCandidate, '-')
		case "meta":
			buf.AppendString(v.Meta, '+')
		}
	}
	return v.Prefix + string(buf), nil
}

func (v Version) String() string {
	result, err := v.Format(FullFormat)
	if err != nil {
		return ""
	}
	return result
}

// PreRelease formats the pre-release version depending on the number n of commits since the
// last tag. If n is zero it returns the parsed pre-release version. If n is greater than zero
// it will append the string "dev.<n>" to the pre-release version.
func (v Version) PreRelease() string {
	if v.Commits == 0 {
		return v.preRelease
	}
	if v.preRelease == "" {
		return fmt.Sprintf("dev.%d", v.Commits)
	}
	return fmt.Sprintf("%s.dev.%d", v.preRelease, v.Commits)
}

func (v Version) ReleaseCandidate() (string, error) {
	if v.preRelease == "" {
		return "rc.1", nil
	}
	re := regexp.MustCompile(`^([a-z]+)\.([0-9]+)$`)
	if !re.MatchString(v.preRelease) {
		return "", errors.New("pre-release does not match the release-candidate format (rc.1, other.1)")
	}
	st := re.FindStringSubmatch(v.preRelease)
	if len(st) != 3 {
		return "", errors.New("pre-release does not match the release-candidate format (rc.1, other.1)")
	}
	i, err := strconv.ParseInt(st[2], 10, 64);
	if err != nil {
		return "", err
	}
	if v.Commits > 0 {
		i++
	}
	return fmt.Sprintf("%s.%d", st[1], i), nil
}

func NewFromHead(head *RepoHead) (Version, error) {
	v := Version{Commits: head.CommitsSinceTag}
	if strings.HasPrefix(head.LastTag, DefaultPrefix) {
		v.Prefix = DefaultPrefix
	}
	version := strings.TrimPrefix(head.LastTag, v.Prefix)
	if strings.Contains(version, "+") {
		parts := strings.Split(version, "+")
		version = parts[0]
		v.Meta = parts[1]
	} else if head.CommitsSinceTag > 0 {
		v.Meta = head.Hash[:8]
	}
	if strings.Contains(version, "-") {
		parts := strings.Split(version, "-")
		version = parts[0]
		v.preRelease = parts[1]
	}

	if version == "" {
		v.Major = 0
		v.Minor = 0
		v.Patch = 0
		return v, nil
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("git version tag must contain 3 components: X.Y.Z: Got %s", version)
	}
	var err error
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return v, fmt.Errorf("failed to parse major version: %v", err)
	}
	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return v, fmt.Errorf("failed to parse minor version: %v", err)
	}
	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return v, fmt.Errorf("failed to parse patch version: %v", err)
	}
	return v, nil
}

// NewFromRepo calculates a semantic version for the head commit of the repo at path.
// If the latest commit is not tagged, the version will have a pre-release-suffix
// appended to it (e.g.: 1.2.3-dev.3+fcf2c8f). The suffix has the format dev.<n>+<hash>,
// whereas n is the number of commits since the last tag and hash is the commit hash
// of the latest commit. NewFromRepo will also increment the patch-level version component
// in case it detects that the current version is a pre-release.
// If the last tag has itself a pre-release-identifier and the last commit is not tagged,
// NewFromRepo will not increment the patch-level version.
// The not SemVer commpliant but commonly used prefix v will be automatically detected.
func NewFromRepo(path string) (Version, error) {
	head, err := GitDescribe(path)
	if err != nil {
		return Version{}, err
	}
	v, err := NewFromHead(head)
	return v, err
}
