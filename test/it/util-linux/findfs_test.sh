TestFindfsMissing() { findfs LABEL=no_such_label_xyz 2>/dev/null; echo "rc=$?"; }
TestFindfsBadSpec() { findfs notatag 2>/dev/null; echo "rc=$?"; }
