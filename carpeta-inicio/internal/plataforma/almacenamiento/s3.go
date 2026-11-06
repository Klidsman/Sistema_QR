package almacenamiento

import (
	"context"
	"errors"
	"fmt"
	"io"

	"asistencia/internal/plataforma/configuracion"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// regionAlmacenamiento se usa con Cloudflare R2. Backblaze B2 acepta el mismo valor con su endpoint.
const regionAlmacenamiento = "auto"

// AlmacenS3 guarda los objetos en un bucket compatible con S3 (Cloudflare R2 o Backblaze B2).
type AlmacenS3 struct {
	cliente *s3.Client
	bucket  string
}

// NuevoAlmacenS3 crea el cliente con las credenciales y el endpoint de la configuración.
func NuevoAlmacenS3(cfg configuracion.Almacenamiento) *AlmacenS3 {
	cliente := s3.New(s3.Options{
		Region:       regionAlmacenamiento,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.ClaveAcceso, cfg.ClaveSecreta, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
	})
	return &AlmacenS3{cliente: cliente, bucket: cfg.Bucket}
}

// Guardar sube el contenido al bucket con la clave indicada.
func (a *AlmacenS3) Guardar(ctx context.Context, clave string, contenido io.Reader) error {
	_, err := a.cliente.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(clave),
		Body:   contenido,
	})
	if err != nil {
		return fmt.Errorf("subiendo el objeto %s: %w", clave, err)
	}
	return nil
}

// Obtener descarga el objeto para lectura.
func (a *AlmacenS3) Obtener(ctx context.Context, clave string) (io.ReadCloser, error) {
	salida, err := a.cliente.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(clave),
	})
	var noEncontrado *types.NoSuchKey
	if errors.As(err, &noEncontrado) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("descargando el objeto %s: %w", clave, err)
	}
	return salida.Body, nil
}

// Eliminar borra el objeto del bucket.
func (a *AlmacenS3) Eliminar(ctx context.Context, clave string) error {
	_, err := a.cliente.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(clave),
	})
	if err != nil {
		return fmt.Errorf("eliminando el objeto %s: %w", clave, err)
	}
	return nil
}
