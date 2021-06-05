package entity

// Activity represents an event made by the user such as `upload`.
type Activity struct {
	Timestamp int64       `json:"timestamp,omitempty"`
	Type      string      `json:"type,omitempty"`
	Content   interface{} `json:"content,omitempty"`
}

type Submission struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Sha256    string `json:"sha256,omitempty"`
}

type Comment struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Sha256    string `json:"sha256,omitempty"`
	Body      string `json:"body,omitempty"`
	ID        string `json:"id,omitempty"`
}

// User represent a user.
type User struct {
	Email            string       `json:"email,omitempty"`
	Username         string       `json:"username,omitempty"`
	Password         string       `json:"password,omitempty"`
	Name             string       `json:"name,omitempty"`
	Location         string       `json:"location,omitempty"`
	URL              string       `json:"url,omitempty"`
	Bio              string       `json:"bio,omitempty"`
	Confirmed        bool         `json:"confirmed,omitempty"`
	MemberSince      int64        `json:"member_since,omitempty"`
	LastSeen         int64        `json:"last_seen,omitempty"`
	Admin            bool         `json:"admin,omitempty"`
	HasAvatar        bool         `json:"has_avatar,omitempty"`
	Following        []string     `json:"following,omitempty"`
	FollowingCount   int          `json:"following_count"`
	Followers        []string     `json:"followers,omitempty"`
	FollowersCount   int          `json:"followers_count"`
	Likes            []string     `json:"likes,omitempty"`
	LikesCount       int          `json:"likes_count"`
	Activities       []Activity   `json:"activities,omitempty"`
	Submissions      []Submission `json:"submissions"`
	SubmissionsCount int          `json:"submissions_count"`
	Comments         []Comment    `json:"comments,omitempty"`
	CommentsCount    int          `json:"comments_count"`
}
