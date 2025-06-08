This folder contains a number of files that are used to generate Go files using Python utilities.

Most of the time, there won't be any need to run any of these files, their output included in the parent directory.

The easiest way to use these scripts is to make sure you have `uv` installed, and then you can simply use `uv run <script.name>` to run the script in question.

(If you know Python, feel free to use `venv` and `pip` (or similar tools) to install `mpmath` (the only dependency) if you prefer, but you might want to look into `uv`, it's VERY fast and easy to use.)

## Constants

`constgen.py` should be used generates the file `constants.go`

```
> uv run constgen.py > ../constants.go
```

This script uses mpmath to generate very high precision constants (most notably π and its multiples, and ln(2)) as hex constants to avoid any run-time arithmetic in the Go code.

It's unlikely that you'll need to modify or run this script unless you are doing major surgery on the transcendental functions.

## Tests

Most of the scripts in this folder are used to generate test data, and are the correct mechanism adding new test cases. These scripts generate the contents of the `fix64_testdata.go` and `fix128_testdata.go` files.

The basic usage is:
   uv run fix64_testgen.py > ../fix64_testdata.go
   uv run fix128_testgen.py > ../fix128_testdata.go

If you want to add new test cases, the format of the input data should be pretty self explanatory. The Python code itself computes the correct answer (using the high precision `Decimal` and `mpmath` libraries). It's important that the tests are _bit accurate_ to ensure that the output of these libraries are identical on all hardware platforms and operating systems.