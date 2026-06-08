TestSpeakerNoText() {
    speaker 2>&1
    echo "rc:$?"
}
