package yadiskapi

import (
	"fmt"
	"time"
)

type SuccessResponse struct {
	OperationId string `json:"operation_Id"`
	Href        string `json:"href"`
	Method      string `json:"method"`
	Templated   bool   `json:"templated"`
}

func (e *SuccessResponse) Info() string {
	return fmt.Sprintf("OperationId: %s\nHref: %s\nMethod: %s\nTemplated: %s\n", e.OperationId, e.Href, e.Method, e.Templated)
}

type ErrorResponse struct {
	Message     string `json:"message"`     //Человекочитаемое описание ошибки
	Description string `json:"description"` //Техническое описание ошибки
	Error       string `json:"error"`       //Уникальный код ошибки
}

func (e *ErrorResponse) Info() string {
	return fmt.Sprintf("message: %s\ndescription: %s\nerror: %s\n", e.Message, e.Description, e.Error)
}

type Disk struct {
	MaxFileSize                int           `json:"max_file_size"`
	PaidMaxFileSize            int64         `json:"paid_max_file_size"`
	TotalSpace                 int64         `json:"total_space"`
	TrashSize                  int           `json:"trash_size"`
	IsPaid                     bool          `json:"is_paid"`
	UsedSpace                  int64         `json:"used_space"`
	SystemFolders              SystemFolders `json:"system_folders"`
	User                       User          `json:"user"`
	UnlimitedAutouploadEnabled bool          `json:"unlimited_autoupload_enabled"`
	Revision                   int64         `json:"revision"`
}

func (d *Disk) String() string {
	return fmt.Sprintf("DISK:\nUser login: %s\nUser UID: %s\nUser country: %s", d.User.Login, d.User.UID, d.User.Country)
}

type User struct {
	RegTime     string `json:"reg_time"`
	DisplayName string `json:"display_name"`
	UID         string `json:"uid"`
	Country     string `json:"country"`
	Login       string `json:"login"`
}

type SystemFolders struct {
	Odnoklassniki string `json:"odnoklassniki"`
	Google        string `json:"google"`
	Instagram     string `json:"instagram"`
	Vkontakte     string `json:"vkontakte"`
	Mailru        string `json:"mailru"`
	Downloads     string `json:"downloads"`
	Applications  string `json:"applications"`
	Facebook      string `json:"facebook"`
	Social        string `json:"social"`
	Scans         string `json:"scans"`
	Screenshots   string `json:"screenshots"`
	Photostream   string `json:"photostream"`
}

type ResourceList struct {
	Embedded struct {
		Sort   string     `json:"sort"`   //Поле, по которому отсортирован список
		Items  []Resource `json:"items"`  //Элементы списка
		Limit  int        `json:"limit"`  //Количество элементов на странице
		Offset int        `json:"offset"` //Смещение от начала списка
		Path   string     `json:"path"`   //Путь к ресурсу, для которого построен список
		Total  int        `json:total`    //Общее количество элементов в списке
	} `json:"_embedded"`
}

func (r *ResourceList) String() string {
	res := "FILES:\n"
	for _, file := range r.Embedded.Items {
		res += file.Path + "\n"
	}
	res += fmt.Sprintf("Items total: %d", r.Embedded.Total)
	return res
}

type TrashResourceList struct {
	Embedded struct {
		Sort   string     `json:"sort"`
		Items  []Resource `json:"items"`
		Limit  int        `json:"limit"`
		Offset int        `json:"offset"`
		Path   string     `json:"path"`
		Total  int        `json:"total"`
	} `json:"_embedded"`
}

// func (t *TrashResourceList) Info() string {
// 	result := ""
// 	for _, file := range t.Embedded.Items {
// 		result += t.Embedded.Path + file.Name + "\n"
// 	}
// 	result += fmt.Sprintf("files count: %d", t.Embedded.Total)

// 	return result
// }

type Resource struct {
	AntivirusStatus  string     `json:"antivirus_status"`   //Статус проверки антивирусом
	ResourceID       string     `json:"resource_id"`        //Идентификатор ресурса
	Share            ShareInfo  `json:"share "`             //
	File             string     `json:"file"`               //URL для скачивания файла
	Size             int        `json:"size"`               //Размер файла
	Photoslice_time  string     `json:"photoslice_time"`    //Дата создания фото или видео файла
	Exif             Exif       `json:"exif,omitempty"`     //
	CustomProperties string     `json:"custom_properties "` //Пользовательские атрибуты ресурса
	MediaType        string     `json:"media_type"`         //Определённый Диском тип файла
	Preview          string     `json:"preview"`            //URL превью файла
	Type             string     `json:"type"`               //Тип
	MimeType         string     `json:"mime_type"`          //MIME-тип файла
	Revision         int64      `json:"revision"`           //Ревизия Диска в которой этот ресурс был изменён последний раз
	PublicUrl        string     `json:"public_Url"`         //Публичный URL
	Path             string     `json:"path"`               //Путь к ресурсу
	Md5              string     `json:"md5"`                //MD5-хэш
	PublicKey        string     `json:"public_Key"`         //Ключ опубликованного ресурса
	Sha256           string     `json:"sha256"`             //SHA256-хэш
	Name             string     `json:"name"`               //Имя
	Created          string     `json:"created"`            //Дата создания
	Sizes            string     `json:"sizes"`              //
	Modified         string     `json:"modified"`           //Дата изменения
	CommentIds       CommentIds `json:"comment_ids"`        //
}

type ShareInfo struct {
	IsRoot  bool   `json:"is_root"`  //Признак того, что папка является корневой в группе
	IsOwned bool   `json:"is_owned"` //Признак, что текущий пользователь является владельцем общей папки
	Rights  string `json:"rights"`   //Права доступа
}

type Exif struct {
	DateTime time.Time `json:"date_time"`
}

type CommentIds struct {
	PrivateResource string `json:"private_resource"` //Идентификатор комментариев для приватных ресурсов
	PublicResource  string `json:"public_resource"`  //Идентификатор комментариев для публичных ресурсов
}
