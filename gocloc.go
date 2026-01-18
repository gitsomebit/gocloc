package gocloc

import (
	"runtime"
	"sync"
)

// Processor is gocloc analyzing processor.
type Processor struct {
	langs *DefinedLanguages
	opts  *ClocOptions
}

// Result defined processing result.
type Result struct {
	Total         *Language
	Files         map[string]*ClocFile
	Languages     map[string]*Language
	MaxPathLength int
}

type job struct {
	file string
	lang *Language
}
type res struct {
	file string
	lang *Language
	cf   *ClocFile
}

// NewProcessor returns Processor.
func NewProcessor(langs *DefinedLanguages, options *ClocOptions) *Processor {
	return &Processor{
		langs: langs,
		opts:  options,
	}
}

// Analyze executes gocloc parsing for the directory of the paths argument and returns the result.
func (p *Processor) Analyze(paths []string) (*Result, error) {
	total := NewLanguage("TOTAL", []string{}, [][]string{{"", ""}})
	languages, err := getAllFiles(paths, p.langs, p.opts)
	if err != nil {
		return nil, err
	}
	maxPathLen := 0
	num := 0
	for _, lang := range languages {
		num += len(lang.Files)
		for _, file := range lang.Files {
			l := len(file)
			if maxPathLen < l {
				maxPathLen = l
			}
		}
	}
	clocFiles := make(map[string]*ClocFile, num)

	jobs := make(chan job, 1024)
	results := make(chan res, 1024)

	workerCount := max(runtime.GOMAXPROCS(0), 1)
	// Avoid opening too many files at once on huge repos
	const maxWorkers = 32
	if workerCount > maxWorkers {
		workerCount = maxWorkers
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				cf := AnalyzeFile(j.file, j.lang, p.opts)
				cf.Lang = j.lang.Name
				results <- res{file: j.file, lang: j.lang, cf: cf}
			}
		}()
	}

	// Feed jobs then close channel
	go func() {
		for _, language := range languages {
			for _, file := range language.Files {
				jobs <- job{file: file, lang: language}
			}
		}
		close(jobs)
	}()

	// Close results after workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Single-threaded aggregation (no race conditions)
	for r := range results {
		r.lang.Code += r.cf.Code
		r.lang.Comments += r.cf.Comments
		r.lang.Blanks += r.cf.Blanks
		clocFiles[r.file] = r.cf
	}

	// Totals
	for _, language := range languages {
		files := int32(len(language.Files))
		if files <= 0 {
			continue
		}
		total.Total += files
		total.Blanks += language.Blanks
		total.Comments += language.Comments
		total.Code += language.Code
	}

	return &Result{
		Total:         total,
		Files:         clocFiles,
		Languages:     languages,
		MaxPathLength: maxPathLen,
	}, nil
}
