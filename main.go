package main

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/mirokuratczyk/consistent-hashing-evaluation/v2/buraksezer"
	"github.com/mirokuratczyk/consistent-hashing-evaluation/v2/intf"
	"github.com/mirokuratczyk/consistent-hashing-evaluation/v2/mirokuratczyk"
)

type run struct {
	numKeys              int
	relocationPercentage float64
}

type consistentHashingImplementation struct {
	newConsistent func() intf.Consistent
	name          string
}

func main() {

	cs := []consistentHashingImplementation{
		{
			name: "mirokuratczyk",
			newConsistent: func() intf.Consistent {
				return mirokuratczyk.NewConsistent()
			},
		},
		// TODO: test different parameterizatinos of buraksezer
		{
			name: "buraksezer",
			newConsistent: func() intf.Consistent {
				return buraksezer.NewConsistent()
			},
		},
	}

	for _, c := range cs {
		eval(c)
	}
}

// TODO: add these metrics
// - #/% keys not distributed
func eval(impl consistentHashingImplementation) {

	runs := make([]*run, 0)

	members := []intf.Member{}
	for i := 0; i < 1000; i++ {
		member := intf.Member(fmt.Sprintf("node%d.olricmq", i))
		members = append(members, member)
	}

	c := impl.newConsistent()
	for _, member := range members {
		c.Add(member)
	}

	keyCount := 100
	// load := (c.AverageLoad() * float64(keyCount)) / float64(cfg.PartitionCount)
	// fmt.Println("Maximum key count for a member should be around this: ", math.Ceil(load))

	// Random distribution: should change between runs
	// fmt.Println("Random distribution")
	distribution := make(map[string]int)
	key := make([]byte, 4)
	for i := 0; i < keyCount; i++ {
		rand.Read(key)
		member := c.LocateKey(key)
		distribution[member.String()]++
	}
	// for member, count := range distribution {
	// 	fmt.Printf("member: %s, key count: %d, load: %f\n", member, count, float64(count)/float64(keyCount))
	// }

	// Deterministic distribution: should not change between runs
	// fmt.Println("Deterministic distribution")
	distribution = make(map[string]int)
	keyToMember := make(map[string]string)
	for i := 0; i < keyCount; i++ {
		key := fmt.Sprintf("%d", i)
		member := c.LocateKey([]byte(key))
		distribution[member.String()]++
		keyToMember[key] = member.String()
	}
	// for member, count := range distribution {
	// 	fmt.Printf("member: %s, key count: %d, load: %f\n", member, count, float64(count)/float64(keyCount))
	// }

	plotDistribution(distribution, fmt.Sprintf("%s_distribution_pre_new_members.html", impl.name))

	// Test deterministic relocation %

	for numNewMembers := 0; numNewMembers < 20; numNewMembers++ {

		member := intf.Member(fmt.Sprintf("node%d.olricmq.new", numNewMembers))
		c.Add(member)

		distribution = make(map[string]int)
		// keyToMember = make(map[string]string) // messes with relocation calculation
		for i := 0; i < keyCount; i++ {
			key := fmt.Sprintf("%d", i)
			member := c.LocateKey([]byte(key))
			distribution[member.String()]++
			// keyToMember[key] = member.String()
		}

		plotDistribution(distribution, fmt.Sprintf("%s_distribution_post_new_members%d.html", impl.name, numNewMembers))

		// Get the new layout and compare with the previous
		var changed int
		for key, oldMember := range keyToMember {
			member := c.LocateKey([]byte(key))
			if member.String() != oldMember {
				changed++
				// fmt.Printf("%s moved to %s from %s\n", key, oldMember, member)
			}
		}

		relocationPercentage := 100 * float64(changed) / float64(keyCount)
		fmt.Printf("\n%s: %.2f%% of the %d/%d keys are relocated when %d new members are added\n",  impl.name, relocationPercentage, changed, keyCount, numNewMembers+1)

		runs = append(runs, &run{
			numKeys:              keyCount,
			relocationPercentage: relocationPercentage,
		})
	}

	plotRuns(runs, impl.name)
}

func plotRuns(runs []*run, plotName string) {

	items := make([]opts.BarData, len(runs))
	for i := 0; i < len(runs); i++ {
		items[i] = opts.BarData{Value: runs[i].relocationPercentage}
	}

	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    plotName,
		Subtitle: "# new members to %% of key relocations",
	}))

	x := make([]int, len(runs))
	for i := 0; i < len(runs); i++ {
		x[i] = i + 1
	}

	bar.SetXAxis(x).
		AddSeries("%% of keys relocated", items)

	f, _ := os.Create(fmt.Sprintf("%s_bar.html", plotName))
	bar.Render(f)
}

func plotDistribution(distribution map[string]int, plotName string) {
	items := make([]opts.BarData, 0)
	for _, count := range distribution {
		items = append(items, opts.BarData{Value: count})
	}

	bar := charts.NewBar()

	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    plotName,
		Subtitle: "dist",
	}))

	x := make([]int, len(items))
	for i := 0; i < len(items); i++ {
		x[i] = i
	}

	bar.SetXAxis(x).
		AddSeries("number of keys per member", items)

	f, _ := os.Create(plotName)
	bar.Render(f)
}
