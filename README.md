# wpgx

[![License][LicenseBadge]](./LICENSE)
[![Travis][TravisBadge]][Travis]
[![TestCoverage][CodeCovBadge]][CodeCov]
[![GoReportCard][GoReportCardBadge]][GoReportCard]
[![Codebeat][CodebeatBadge]][CodeBeat]
[![GoDoc][GoDocBadge]][GoDoc]

Wrapped PGX - some utility add-on to the perfect https://github.com/jackc/pgx package

Package wpgx helps to improve loading performance and to simplify code.

Package wpgx uses concept of Shaper and Translator interfaces, that helps
to map database names with struct fields.

Imagine, we have a simple User struct with an access Role field:

    type User struct {
        ID   int      `json:"id"`
        Name string   `json:"name"`
        Role UserRole `json:"role"`
    }

    type UserRole struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

It is business-level objects, let's create a database model:

    type userModel struct {
        ID       int
        Name     string
        RoleID   sql.NullInt64
        RoleName sql.NullString
    }

Note that model can have unexported name, it's because we use interfaces in wpgx.

To convert business to database we can implement Extrude method in Shaper interface:

    func (u *User) Extrude() wpgx.Translator {
        return &userModel{
            ID:   u.ID,
            Name: u.Name,
            RoleID: sql.NullInt64{
                Int64: int64(u.Role.ID),
                Valid: u.Role.ID > 0,
            },
            RoleName: sql.NullString{
                String: u.Role.Name,
                Valid:  u.Role.Name != "",
            },
        }
    }

Okay, to convert from database to business we can implement Receive method in Shaper interface:

    func (u *User) Receive(item wpgx.Translator) error {
        m, ok := item.(*userModel)
        if !ok {
            return wpgx.ErrUnknownType
        }

        u.ID = m.ID
        u.Name = m.Name

        // No role selected, set empty
        if !m.RoleID.Valid {
            u.Role = UserRole{}
            return nil
        }

        u.Role.ID = int(m.RoleID.Int64)

        if m.RoleName.Valid {
            u.Role.Name = m.RoleName.String
        }
        return nil
    }

Now, we can use something like this query to fetch user:

    SELECT
        u.id,
        u.name,
        u.role_id,
        r.name as role_name
    FROM users u
    LEFT JOIN roles r ON u.role_id = r.id
    WHERE u.id = $1;

Let's implement Translator interface to map names:

    func (um *userModel) Translate(name string) interface{} {
        switch name {
        case "id":
            return &um.ID
        case "name":
            return &um.Name
        case "role_id":
            return &um.RoleID
        case "role_name":
            return &um.RoleName
        }
        return nil
    }

To create database connection, just call Connect:

    db, err := wpgx.Connect("postgresql://user:pass@host:port/database?options")
    if err != nil {
        return err
    }

For performance issues we can to prepare statements but it is not required for select:

    sqlSelectUser, err := db.Cook(`query text`)
    if err != nil {
        return err
    }

Now we can load user with ID=42:

    var user User

    if err = db.Load(&user, sqlSelectUser, 42); err != nil {
        return err
    }

Okay, but what if we need a list of users? No problems, let's implement Collector interface:

    type UserList []*User

    func (ul *UserList) NewItem() wpgx.Shaper {
        return new(User)
    }

    func (ul *UserList) Collect(item Shaper) error {
        user, ok := item.(*User)
        if !ok {
            return wpgx.ErrUnknownType
        }
        *ul = append(*ul, user)
        return nil
    }

Now we can select list like this. For example all users with role:

    sqlSelectUsers, err := db.Cook(`
    SELECT
        u.id,
        u.name,
        u.role_id,
        r.name as role_name
    FROM users u
    LEFT JOIN roles r ON u.role_id = r.id
    WHERE u.role_id IS NOT NULL LIMIT 100;
    `)
    if err != nil {
        return err
    }

    users := make(UserList, 0, 100)

    err = db.Deal(&users, sqlSelectUsers)
    if err != nil {
        return err
    }

Collector can use maps, channels or another type you want.

But what if we need to insert new user and fetch this id?
For this task we have to prepare insert query with columns:

    sqlInsertUser, err := db.Cook(`
    INSERT INTO users (name, role_id)
    VALUES ($1, $2)
    RETURNING id;
    `, "name", "role_id") // describe Translator columns for save
    if err != nil {
        return err
    }

    var ids wpgx.Ints
    user := &User{Name: "John"}

    if err = db.Save(user, sqlInsertUser, &ids); err != nil {
        return err
    }
    if len(ids) > 0 {
        user.ID = ids[0]
    }

Save uses Collector as result type because you may want to return many rows.

All examples above can be used in one transaction, that called Dealer.
Typically Dealer can be used like this:

    d, err := db.NewDealer()
    if err != nil {
        return err
    }

    // Commit when no errors happened
    defer func(){ d.Jail(err == nil) }()

    // Doing something good stuff

Good luck!

[Travis]: https://travis-ci.org/shestakovda/wpgx
[CodeCov]: https://codecov.io/gh/shestakovda/wpgx
[GoReportCard]: https://goreportcard.com/report/github.com/shestakovda/wpgx
[CodeBeat]: https://codebeat.co/badges/4238a10d-158b-4116-aac1-b7e21799d8c1
[GoDoc]: https://godoc.org/github.com/shestakovda/wpgx

[LicenseBadge]: https://img.shields.io/dub/l/vibe-d.svg
[TravisBadge]: https://travis-ci.org/shestakovda/wpgx.svg?style=flat-square&&branch=master
[CodeCovBadge]: https://codecov.io/gh/shestakovda/wpgx/branch/master/graph/badge.svg
[GoReportCardBadge]: https://goreportcard.com/badge/github.com/shestakovda/wpgx
[CodebeatBadge]: https://codebeat.co/projects/github-com-shestakovda-wpgx-master
[GoDocBadge]: https://godoc.org/github.com/shestakovda/wpgx?status.svg