package profile

import (
	"backend/database/ent"
	"backend/database/ent/user"
	"backend/file"
	"context"

	"github.com/cockroachdb/errors"
	"go.uber.org/fx"
)

type Profile struct {
	file file.File
}

type profileParams struct {
	fx.In
	File file.File
}

func newProfile(p profileParams) *Profile {
	return &Profile{
		file: p.File,
	}
}

func (profile *Profile) GetProfile(ctx context.Context, client *ent.Client, userId int) (*ent.User, error) {
	user, err := client.User.Query().Where(user.IDEQ(userId)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "User.Query() failed")
	}
	if user == nil {
		return nil, errors.New("User not found")
	}

	return user, nil
}

func (profile *Profile) UpdateProfile(ctx context.Context, client *ent.Client, p *UpdateProfileParams) (*ent.User, error) {
	user, err := client.User.Query().Where(user.IDEQ(p.UserId)).First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "User.Query() failed")
	}
	if user == nil {
		return nil, errors.New("User not found")
	}

	updateBuilder := user.Update()
	if p.Fullname != "" {
		updateBuilder.SetFullname(p.Fullname)
	}
	if p.Phone != "" {
		updateBuilder.SetPhone(p.Phone)
	}
	if p.Avatar != "" {
		avatar, err := profile.file.MoveFromTemporaryAndDeleteOldFile(p.Avatar, file.FolderUser, user.Avatar)
		if err != nil {
			return nil, err
		}
		updateBuilder.SetAvatar(avatar)
	}

	user, err = updateBuilder.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "User.Update() failed")
	}
	return user, nil
}

type UpdateProfileParams struct {
	UserId   int
	Fullname string
	Phone    string
	Avatar   string
}
