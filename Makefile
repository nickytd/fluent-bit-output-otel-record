all: out_gstdout.so

out_gstdout.so:
	go build -buildmode=c-shared -o out_gstdout.so .

fast:
	go build out_gstdout.go

clean:
	rm -rf *.so *.h *~

run: out_gstdout.so
	fluent-bit -c fluent-bit.yaml -e out_gstdout.so