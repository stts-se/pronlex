package decompounder

import (
	"fmt"
	"testing"
)

var ts = "Wanted '%v' got '%v'\n"

func spunk() { fmt.Println() }

func Test_Tree(t *testing.T) {

	tr := NewTNode()

	if want, got := rune(0), tr.r; want != got {
		t.Errorf(ts, want, got)
	}

	tr = tr.add("strut")
	tr = tr.add("strutnos")
	tr = tr.add("strutnosar")

	// for k, v := range tr.sons {
	//     fmt.Printf("HOJSAN: %#v : %s\n", k, string(v.r))
	// }

	if want, got := rune(0), tr.r; want != got {
		t.Errorf(ts, want, got)
	}

	if want, got := 1, len(tr.sons); want != got {
		t.Errorf(ts, want, got)
	}

	s1 := "strutnosarna"
	prfs := tr.prefixes(s1)
	//fmt.Printf("Arcs: %#v\n", prfs)
	if want, got := 3, len(prfs); want != got {
		t.Errorf(ts, want, got)
	}

	// for _, p := range prfs {
	//     fmt.Printf("PREFIX: %v\n", s1[p.start:p.end])
	// }

	pt := NewPrefixTree()
	pt.Add("ap")
	pt.Add("hund")
	pt.Add("aphund")
	pt.Add("nos")

	s := "aphundar"
	arczz := pt.Prefixes(s)
	if want, got := 2, len(arczz); want != got {
		t.Errorf(ts, want, got)
	}

	prefs := map[string]bool{s[arczz[0].start:arczz[0].end]: true, s[arczz[1].start:arczz[1].end]: true}
	if _, ok := prefs["ap"]; !ok {
		t.Errorf(ts, "'ap'", "nothing")
	}
	if _, ok := prefs["aphund"]; !ok {
		t.Errorf(ts, "'aphund'", "nothing")
	}

	st := NewSuffixTree()

	st.Add("rot")
	st.Add("mos")
	st.Add("nos")

	z := "skrotmos"
	arcs := st.Suffixes(z)
	if want, got := 1, len(arcs); want != got {
		t.Errorf(ts, want, got)
	}

	st.Add("rotmos")
	arcs = st.Suffixes(z)
	//fmt.Printf("ARKZ: %#v\n", arcs)
	if want, got := 2, len(arcs); want != got {
		t.Errorf(ts, want, got)
	}

	suffs := map[string]bool{z[arcs[0].start:arcs[0].end]: true, z[arcs[1].start:arcs[1].end]: true}
	if _, ok := suffs["mos"]; !ok {
		t.Errorf(ts, "'mos'", "nothing")
	}
	if _, ok := suffs["rotmos"]; !ok {
		t.Errorf(ts, "'rotmos'", "nothing")
	}

}

func Test_Paths(t *testing.T) {

	a1 := arc{start: 0, end: 3}
	a2 := arc{start: 3, end: 7}

	res := paths([]arc{a1, a2}, 0, 7)

	if want, got := 1, len(res); want != got {
		t.Errorf("NOOOO! %d %d", want, got)
	}
	p := res[0]
	if want, got := 2, len(p); want != got {
		t.Errorf("AAAA! %d %d", want, got)

	}
	a1_ := p[0]
	if want, got := 0, a1_.start; want != got {
		t.Errorf("AAAA! %d %d", want, got)

	}
	if want, got := 3, p[1].start; want != got {
		t.Errorf("AAAA! %d %d", want, got)

	}

	a3 := arc{start: 3, end: 5}
	a4 := arc{start: 5, end: 7}
	a5 := arc{start: 3, end: 6}

	res = paths([]arc{a1, a2, a3, a4, a5}, 0, 7)
	if want, got := 2, len(res); want != got {
		t.Errorf("Suck %d %d", want, got)
	}
	//fmt.Printf("\n%#v\n", res)
}

func Test_Decompounder(t *testing.T) {

	d := NewDecompounder()

	d.prefixes.Add("sylt")
	d.prefixes.Add("syl")

	d.suffixes.Add("järn")
	d.suffixes.Add("tjärn")

	decomps := d.Decomp("syltjärn")
	if w, g := 2, len(decomps); w != g {
		t.Errorf(ts, w, g)
	}
	if w, g := 2, len(decomps[0]); w != g {
		t.Errorf(ts, w, g)
	}
	if w, g := 2, len(decomps[1]); w != g {
		t.Errorf(ts, w, g)
	}

	p1 := decomps[0][0]
	p2 := decomps[1][0]

	if p1 == p2 {
		t.Error("Aouch")
	}

	if p1 != "syl" && p2 != "syl" {
		t.Error("Aouch")
	}
	if p1 != "sylt" && p2 != "sylt" {
		t.Error("Aouch")
	}
}
