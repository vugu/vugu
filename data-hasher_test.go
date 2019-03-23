package vugu

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHash(t *testing.T) {

	assert := assert.New(t)

	data := struct {
		FString    string
		FInt       int
		FFloat     float64
		FMap       map[string]bool
		FSlice     []string
		FNil       interface{}
		FStruct    struct{ Inner1 string }
		FStructPtr *bytes.Buffer
		unexported bool
	}{
		FString:    "string1",
		FInt:       10,
		FFloat:     10.0,
		FMap:       map[string]bool{"key1": true, "key2": false},
		FSlice:     []string{"Larry", "Moe", "Curly"},
		unexported: true,
	}

	// log.Printf("1-----")
	lasth := ComputeHash(&data)
	// log.Printf("2-----")
	assert.Equal(lasth, ComputeHash(&data))
	// log.Printf("3-----")
	lasth = ComputeHash(data)
	assert.Equal(lasth, ComputeHash(data))
	data.FString = "string2"
	assert.NotEqual(lasth, ComputeHash(data))
	data.FString = "string1"
	assert.Equal(lasth, ComputeHash(data))
	data.FMap = nil
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.unexported = false
	assert.Equal(lasth, ComputeHash(data))
	data.FStruct.Inner1 = "someval"
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.FStructPtr = &bytes.Buffer{}
	assert.NotEqual(lasth, ComputeHash(data))
	lasth = ComputeHash(data)
	data.FNil = "not nil any more"
	assert.NotEqual(lasth, ComputeHash(data))

	// log.Printf("HERE1")
	data.FMap = map[string]bool{"key1": true, "key2": false}
	lasth = ComputeHash(data)
	data.FMap["key2"] = true
	assert.NotEqual(lasth, ComputeHash(data))
	data.FMap["key2"] = false
	assert.Equal(lasth, ComputeHash(data))
	data.FMap["key3"] = true
	assert.NotEqual(lasth, ComputeHash(data))
	delete(data.FMap, "key3")
	assert.Equal(lasth, ComputeHash(data))

}

func BenchmarkComputeHash(b *testing.B) {

	b.StopTimer()

	data := struct {
		FString    string
		FInt       int
		FFloat     float64
		FMap       map[string]bool
		FSlice     []string
		FNil       interface{}
		FStruct    struct{ Inner1 string }
		FStructPtr *bytes.Buffer
		unexported bool
	}{
		FString:    "string1",
		FInt:       10,
		FFloat:     10.0,
		FMap:       map[string]bool{"key1": true, "key2": false},
		FSlice:     []string{"Larry", "Moe", "Curly"},
		unexported: true,
	}

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ComputeHash(&data) // 2us on my laptop - obviously data is small but yeah... that's fast
	}

}

type bmNewsItem struct {
	Title string
	Desc  string
}

// ~3k of data in bmNewsSample
var bmNewsSample = []bmNewsItem{
	bmNewsItem{
		Title: "Saudi-Led coalition launches raids on Yemen's Houthis in Sanaa: Al-Arabiya TV",
		Desc:  "The Saudi-Led coalition in Yemen launched raids on Houthi camps in the capital Sanaa, including the Al-Dailami air base, Al-Arabiya TV reported on Saturday.",
	},

	bmNewsItem{
		Title: "China rescuers pull survivor from blast rubble as death toll rises",
		Desc:  "Rescuers pulled a survivor from rubble early on Saturday in the wake of a massive explosion at a pesticide plant in eastern China that flattened buildings, blew out windows more than a mile away and killed at least 64 people.",
	},

	bmNewsItem{
		Title: "Many 'march for love' in New Zealand as mosques reopen",
		Desc:  "About 3,000 people walked through Christchurch in a 'march for love' early on Saturday, honoring the 50 worshippers massacred in the New Zealand city a week ago, as the mosques where the shooting took place reopened for prayers.",
	},

	bmNewsItem{
		Title: "Venezuela national soccer coach offers to resign after meeting Guaido envoy",
		Desc:  "The coach of Venezuela's national soccer team said on Friday he had offered to resign after meeting with opposition leader Juan Guaido's envoy to Spain, in the midst of a power struggle between Guaido and socialist President Nicolas Maduro.",
	},

	bmNewsItem{
		Title: "Trump decides against more North Korea sanctions at this time: source",
		Desc:  "U.S. President Donald Trump on Friday said he has decided against imposing new large-scale sanctions on North Korea in a confusing tweet that seemed to imply he was reversing measures against two Chinese shipping companies, a U.S. administration source familiar with the matter said.",
	},

	bmNewsItem{
		Title: "Cholera cases reported as hunger, disease stalk African cyclone survivors",
		Desc:  "Cholera cases were reported on Friday in the Mozambican city of Beira, adding a risk of deadly illnesses for hundreds of thousands of people who are scrambling for shelter, food and water after catastrophic flooding in southern Africa.",
	},

	bmNewsItem{
		Title: "Trump did not reverse North Korea sanctions on Chinese shipping companies: source",
		Desc:  "U.S. President Donald Trump on Friday did not reverse North Korea-related sanctions on two Chinese shipping companies, a U.S. administration source familiar with the matter said.",
	},

	bmNewsItem{
		Title: "China chemical plant blast kills 62; Xi orders probe",
		Desc:  "An explosion at a pesticide plant in eastern China has killed 62 people, state media said late on Friday, with 34 people critically injured and 28 missing, the latest casualties in a series of industrial accidents that has angered the public.",
	},

	bmNewsItem{
		Title: "Trump, Germany's Merkel discuss trade, NATO funding, Brexit",
		Desc:  "U.S. President Donald Trump spoke by phone with German Chancellor Angela Merkel on Friday to discuss a range of issues including trade and NATO funding, the White House and a senior administration official said.",
	},

	bmNewsItem{
		Title: "U.S. lawmaker seeks Boeing whistleblowers, some MAX 737 orders in jeopardy",
		Desc:  "A U.S. lawmaker on Friday urged current or former Boeing Co and Federal Aviation Administration (FAA) employees to come forward with any information about the certification program for the 737 MAX, which has suffered two fatal crashes in five months.",
	},
}

func BenchmarkComputeHashBigTime(b *testing.B) {

	b.StopTimer()

	var a []bmNewsItem

	// benchmark with a tree of about 10MB of data
	for i := 0; i < 3333; i++ {
		a = append(a, bmNewsSample...)
	}

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ComputeHash(a) // 12ms on my laptop - pretty good...
	}

}
