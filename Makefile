BIN_DIR ?= bin

.PHONY: wfg
wfg:
	go build -o $(BIN_DIR)/wfg cmd/waveformgenerator/main.go 

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
