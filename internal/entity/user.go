package entity

type User struct {
	Id         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Username   string    `gorm:"column:username"`
	Name       string    `gorm:"column:name"`
	Email      string    `gorm:"column:email"`
	Password   string    `gorm:"column:password"`
	AvatarPath string    `gorm:"column:avatar_path"`
	Auditable  Auditable `gorm:"embedded"`
}

func (userEntity *User) GetAuditable() *Auditable {
	return &userEntity.Auditable
}
