package daemon

import (
	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"github.com/zmb3/spotify/v2"
)

type Handler struct {
	Commander *commands.Commander
}

type PlayArgs struct{}

func (h *Handler) Play(args *PlayArgs, reply *string) error {
	return h.Commander.Play()
}

type PlayURLArgs struct {
	URL string
}

func (h *Handler) PlayURL(args *PlayURLArgs, reply *string) error {
	return h.Commander.PlayURL(args.URL)
}

type PauseArgs struct{}

func (h *Handler) Pause(args *PauseArgs, reply *string) error {
	return h.Commander.Pause()
}

type TogglePlayArgs struct{}

func (h *Handler) TogglePlay(args *TogglePlayArgs, reply *string) error {
	return h.Commander.TogglePlay()
}

type LinkArgs struct{}

func (h *Handler) Link(args *LinkArgs, reply *string) error {
	link, err := h.Commander.PrintLink()
	*reply = link
	return err
}

type LinkContextArgs struct{}

func (h *Handler) LinkContext(args *LinkContextArgs, reply *string) error {
	link, err := h.Commander.PrintLinkContext()
	*reply = link
	return err
}

type YoutubeLinkArgs struct{}

func (h *Handler) YoutubeLink(args *YoutubeLinkArgs, reply *string) error {
	link, err := h.Commander.PrintYoutubeLink()
	*reply = link
	return err
}

type NextArgs struct {
	Amount int
}

func (h *Handler) Next(args *NextArgs, reply *string) error {
	return h.Commander.Next(args.Amount, false)
}

type PreviousArgs struct{}

func (h *Handler) Previous(args *PreviousArgs, reply *string) error {
	return h.Commander.Previous()
}

type LikeArgs struct{}

func (h *Handler) Like(args *LikeArgs, reply *string) error {
	return h.Commander.Like()
}

type UnlikeArgs struct{}

func (h *Handler) Unlike(args *UnlikeArgs, reply *string) error {
	return h.Commander.UnLike()
}

type NowPlayingArgs struct {
	Force bool
}

func (h *Handler) NowPlaying(args *NowPlayingArgs, reply *string) error {
	resp, err := h.Commander.NowPlaying(args.Force)
	*reply = resp
	return err
}

type VolumeArgs struct {
	Amount int
}

func (h *Handler) ChangeVolume(args *VolumeArgs, reply *string) error {
	return h.Commander.ChangeVolume(args.Amount)
}

type MuteArgs struct{}

func (h *Handler) Mute(args *MuteArgs, reply *string) error {
	return h.Commander.Mute()
}

type UnmuteArgs struct{}

func (h *Handler) Unmute(args *UnmuteArgs, reply *string) error {
	return h.Commander.UnMute()
}

type ToggleMuteArgs struct{}

func (h *Handler) ToggleMute(args *ToggleMuteArgs, reply *string) error {
	return h.Commander.ToggleMute()
}

type DownloadCoverArgs struct {
	Path string
}

func (h *Handler) DownloadCover(args *DownloadCoverArgs, reply *string) error {
	return h.Commander.DownloadCover(args.Path)
}

type RadioArgs struct{}

func (h *Handler) Radio(args *RadioArgs, reply *string) error {
	return h.Commander.Radio()
}

type ClearRadioArgs struct{}

func (h *Handler) ClearRadio(args *ClearRadioArgs, reply *string) error {
	return h.Commander.ClearRadio()
}

type RefillRadioArgs struct{}

func (h *Handler) RefillRadio(args *RefillRadioArgs, reply *string) error {
	return h.Commander.RefillRadio()
}

type StatusArgs struct{}

func (h *Handler) Status(args *StatusArgs, reply *string) error {
	status, err := h.Commander.Status()
	*reply = status
	return err
}

type ListDevicesArgs struct{}

func (h *Handler) Devices(args *ListDevicesArgs, reply *string) error {
	devices, err := h.Commander.ListDevices()
	*reply = devices
	return err
}

type SetDeviceArgs struct {
	DeviceID spotify.ID
}

func (h *Handler) SetDevice(args *SetDeviceArgs, reply *string) error {
	return h.Commander.SetDevice(args.DeviceID)
}

type RepeatArgs struct{}

func (h *Handler) Repeat(args *RepeatArgs, reply *string) error {
	return h.Commander.Repeat()
}

type ShuffleArgs struct{}

func (h *Handler) Shuffle(args *ShuffleArgs, reply *string) error {
	return h.Commander.Shuffle()
}

type SetPositionArgs struct {
	Position int
}

func (h *Handler) SetPosition(args *SetPositionArgs, reply *string) error {
	return h.Commander.SetPosition(args.Position)
}

type SeekArgs struct {
	Fwd bool
}

func (h *Handler) Seek(args *SeekArgs, reply *string) error {
	return h.Commander.Seek(args.Fwd)
}
