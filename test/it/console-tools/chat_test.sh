# chat runs an expect/send conversation over stdin/stdout; a pipe stands in for
# a serial link. Feeding it the expected string makes it send the reply.
TestChatSendsReply() { printf 'OK' | chat OK GO | grep -c 'GO'; }
# A script is required.
TestChatNoScript() { chat </dev/null 2>/dev/null; echo "rc=$?"; }
# An expect string that never arrives is a deterministic failure (EOF).
TestChatExpectNeverSeen() { printf 'nope' | chat LOGIN: user 2>/dev/null; echo "rc=$?"; }
