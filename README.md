# gopi-input

This respository contains input and keymap modules for gopi. It
supports keyboards, mice and touchscreens at present. The gopi
modules provided are:

| Platform | Import | Type | Name |
| -------- | ------ | ---- | ---- |
| Linux    | `github.com/djthorpe/gopi-input/sys/input`  | `gopi.MODULE_TYPE_INPUT` | `linux/input` |
| Any      | `github.com/djthorpe/gopi-input/sys/keymap` | `gopi.MODULE_TYPE_KEYMAP` | `sys/keymap` |

The `input` module provides an Input Manager which can be used for discovering input
devices (keyboards, mouse or touchscreen). It publishes events when the input devices
receive events (key presses, releases and cursor moves, for example).

The `keymap` module can receive input events from keyboards and output runes based
on a set of rules. For example, the 'A' key pressed whilst the shift key is pressed will
result in the upper-case 'A' rune event being published, and so forth. You can
create, modify and delete keymap files through this module.


