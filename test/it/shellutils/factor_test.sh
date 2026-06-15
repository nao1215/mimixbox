# shellcheck shell=sh
# Integration helper for `factor`.
#
# Convention (see test/it/README.md):
#   - All temp state lives under ${MIMIXBOX_IT_ROOT} (per-run root from
#     test/it/spec/spec_helper.sh); never hardcode /tmp/mimixbox.
#   - Setup builds fixtures, CleanUp removes the per-command subdir, and
#     Test<Command><Case> functions exercise real behavior. Idempotent and
#     safe under repeated/parallel runs because everything is under the root.

Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/factor
    export LANG=C
    mkdir -p "${TEST_DIR}"
    # Fixture: whitespace-separated integers read from stdin.
    printf '6 8 12\n' > "${TEST_DIR}/numbers.txt"
}

CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/factor"; }

# Single operand on the command line.
TestFactorArg() { factor 360; }

# Multiple operands.
TestFactorMultipleArgs() { factor 7 12; }

# Reads whitespace-separated numbers from stdin when no operand is given.
TestFactorFromStdin() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/factor
    factor < "${TEST_DIR}/numbers.txt"
}

# A prime factors to itself.
TestFactorPrime() { factor 13; }
