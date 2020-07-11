# CMDS is a list of all folders in cmd/
CMDS = $(notdir $(wildcard cmd/*))
PACKAGE = siody.home/om-like

build/cmd: $(foreach CMD,$(CMDS),build/cmd/$(CMD))

echo:
	echo $(foreach CMD,$(CMDS),build/cmd/$(CMD))

build/cmd/%:
	go build -o $*.o $(PACKAGE)/cmd/$*

clean:
	rm $(foreach CMD,$(CMDS),$(CMD)).o