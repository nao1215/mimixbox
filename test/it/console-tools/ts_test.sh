# Each line gets a YYYY-MM-DD HH:MM:SS prefix; check the shape, not the value.
TestTs() { printf 'alpha\nbeta\n' | ts | grep -cE '^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} (alpha|beta)$'; }
