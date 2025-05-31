The files in this folder should only be needed if new test cases are introduced. They generate the contents of the fix64_testdata.go and fix128_testdata.go files.

The basic usage is:
   python3 fix64_testgen.py > ../fix64_testdata.go

The basic 64 and 128 test generation process doesn't depend on any Python libraries, and should be executable without installing anything with pip or setting up a virtual environment. The transcendental tests do require an external library. If you have access to uv, you can just execute "uv run -m trans_testgen.py" 