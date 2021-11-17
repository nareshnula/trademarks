package main

type Certificate struct {
	SerialNumber       string `json:"serialNumber"`
	RegistrationNumber int    `json:"registrationNumber"`
	RegistrationDate   string `json:"registrationDate"`
	CertificateHash    string `json:"certificateHash"`
	MetaDataHash       string `json:"metaDataHash"`
	PublicationDate    string `json:"publicationDate"`
}
