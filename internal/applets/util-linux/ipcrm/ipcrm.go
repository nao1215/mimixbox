// Package ipcrm implements the ipcrm applet: remove System V IPC objects
// (message queues, shared memory segments, and semaphore arrays) by id.
package ipcrm

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the ipcrm applet.
type Command struct{}

// New returns an ipcrm command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ipcrm" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove System V IPC objects by id" }

// removeIPC is indirected so removal can be tested without real IPC objects.
var removeIPC = func(kind string, id int) error {
	var errno unix.Errno
	switch kind {
	case "msg":
		_, _, errno = unix.Syscall(unix.SYS_MSGCTL, uintptr(id), uintptr(unix.IPC_RMID), 0)
	case "shm":
		_, _, errno = unix.Syscall(unix.SYS_SHMCTL, uintptr(id), uintptr(unix.IPC_RMID), 0)
	case "sem":
		_, _, errno = unix.Syscall(unix.SYS_SEMCTL, uintptr(id), 0, uintptr(unix.IPC_RMID))
	}
	if errno != 0 {
		return errno
	}
	return nil
}

// Run executes ipcrm.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-q msqid] [-m shmid] [-s semid]...", stdio.Err).WithHelp(command.Help{
		Description: "Remove the System V IPC objects with the given ids: -q a message queue, -m a " +
			"shared memory segment, -s a semaphore array. Each option may be given more than once. " +
			"Removal by key (-Q/-M/-S) is not supported.",
		Examples: []command.Example{
			{Command: "ipcrm -q 0 -m 32769", Explain: "Remove a message queue and a shared memory segment."},
		},
		ExitStatus: "0  every object was removed.\n1  an id was invalid or could not be removed.",
	})
	queues := fs.IntSliceP("queue-id", "q", nil, "remove the message queue with this id")
	shmems := fs.IntSliceP("shmem-id", "m", nil, "remove the shared memory segment with this id")
	sems := fs.IntSliceP("semaphore-id", "s", nil, "remove the semaphore array with this id")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if len(*queues)+len(*shmems)+len(*sems) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "ipcrm: nothing to remove (use -q, -m, or -s)")
		return command.SilentFailure()
	}

	failed := false
	remove := func(kind, label string, ids []int) {
		for _, id := range ids {
			if err := removeIPC(kind, id); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "ipcrm: removing %s %d failed: %v\n", label, id, err)
				failed = true
			}
		}
	}
	remove("msg", "queue", *queues)
	remove("shm", "shared memory segment", *shmems)
	remove("sem", "semaphore", *sems)

	if failed {
		return command.SilentFailure()
	}
	return nil
}
