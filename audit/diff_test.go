package audit_test

import (
	. "github.com/nojyerac/go-lib/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompactDiff", func() {
	It("returns nil when both maps are empty", func() {
		Expect(CompactDiff(nil, nil)).To(BeNil())
	})

	It("returns nil when maps are equal", func() {
		before := map[string]any{"status": "new", "count": 1}
		after := map[string]any{"status": "new", "count": 1}
		Expect(CompactDiff(before, after)).To(BeNil())
	})

	It("returns added removed and changed keys", func() {
		before := map[string]any{"status": "new", "count": 1, "remove": true}
		after := map[string]any{"status": "done", "count": 1, "added": "v"}

		diff := CompactDiff(before, after)
		Expect(diff).To(HaveLen(3))
		Expect(diff["status"]).To(Equal(Change{Before: "new", After: "done"}))
		Expect(diff["remove"]).To(Equal(Change{Before: true}))
		Expect(diff["added"]).To(Equal(Change{After: "v"}))
	})
})
