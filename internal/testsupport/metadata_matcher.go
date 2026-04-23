package testsupport

import (
	"fmt"
	"reflect"

	"github.com/lukewilliamboswell/libasciidoc/types"
	"github.com/google/go-cmp/cmp"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/ginkgo/v2"
	gomegatypes "github.com/onsi/gomega/types"
	"github.com/pkg/errors"
)

func MatchMetadata(expected types.Metadata) gomegatypes.GomegaMatcher {
	return &metadataMatcher{
		expected: expected,
	}
}

type metadataMatcher struct {
	expected types.Metadata
	diffs    string
}

func (m *metadataMatcher) Match(actual interface{}) (success bool, err error) {
	actualMeta, ok := actual.(types.Metadata)
	if !ok {
		return false, errors.Errorf("MatchMetadata matcher expects a 'types.Metadata' (actual: %T)", actual)
	}
	// When expected Attributes is nil, treat it as "don't care" —
	// copy the actual value so the comparison ignores it.
	expected := m.expected
	if expected.Attributes == nil {
		expected.Attributes = actualMeta.Attributes
	}
	if !reflect.DeepEqual(expected, actualMeta) {
		GinkgoT().Logf("actual HTML:\n'%s'", actualMeta)
		GinkgoT().Logf("expected HTML:\n'%s'", expected)
		m.diffs = cmp.Diff(spew.Sdump(expected), spew.Sdump(actualMeta))
		return false, nil
	}
	return true, nil
}

func (m *metadataMatcher) FailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected document metadata to match:\n%s", m.diffs)
}

func (m *metadataMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected document metadata not to match:\n%s", m.diffs)
}
