TestIostatCpuHeader() { iostat | sed -n '1p'; }
TestIostatDeviceHeader() { iostat | grep -c 'Device'; }
