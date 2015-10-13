package benchmark

var compilers = map[string]string{
    "C 11":           "1",
    "C++ 11":         "2",
    "delphi (fpc)":   "3",
    "pascal (fpc)":   "4",
    "python2":        "5",
    "python3":        "6",
    "java":           "7",
    "C# (mono dmcs)": "8",
    "C#solution":     "11",
}

func CompilerId(name string) string {
    id, ok := compilers[name]
    if !ok {
        return name
    }
    return id
}
