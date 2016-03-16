package line

func equals(expect map[Field]string, result map[Field]string) bool {
	if len(expect) != len(result) {
		return false
	}
	for f, expS := range expect {
		resS := result[f]
		if resS != expS {
			return false
		}
	}
	return true
}

// Equals compares two line.Format instances
func Equals(x Format, r Format) bool {
	if x.Name != r.Name {
		return false
	}
	if x.FieldSep != r.FieldSep {
		return false
	}
	if x.NFields != r.NFields {
		return false
	}
	if len(x.Fields) != len(r.Fields) {
		return false
	}
	for f, expS := range x.Fields {
		resS := r.Fields[f]
		if resS != expS {
			return false
		}
	}
	return true
}

type stringSlice []string

func (a stringSlice) Len() int      { return len(a) }
func (a stringSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a stringSlice) Less(i, j int) bool {
	return a[i] < a[j]
}