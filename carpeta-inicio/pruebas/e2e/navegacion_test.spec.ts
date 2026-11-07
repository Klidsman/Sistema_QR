import { expect, test } from '@playwright/test';

import { identificadorUnico, ingresarComoAdministrador } from './ayudas';

test.describe('Navegación principal', () => {
  test('el menú lleva a las pantallas sin editar la URL', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const navegacion = page.getByRole('navigation');

    await navegacion.getByRole('link', { name: 'Portería' }).click();
    await expect(page.getByRole('heading', { name: 'Registro de ingreso' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Justificaciones' }).click();
    await expect(page.getByRole('heading', { name: 'Justificaciones pendientes' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Reportes', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Reporte diario' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Resumen', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Reporte acumulado' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Configuración', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Configuración' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Asignaciones', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Asignaciones de docente' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Docentes', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Docentes' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Alumnos', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Alumnos' })).toBeVisible();
  });

  test('al cambiar de administrador a docente llega al inicio docente', async ({ page }) => {
    const sufijo = identificadorUnico();
    const correo = `docente${sufijo}@colegio.pe`;
    const clave = 'clave-docente-2026';

    await ingresarComoAdministrador(page);
    await page.goto('/docentes');
    await page.getByLabel('Nombre completo *').fill(`Docente ${sufijo}`);
    await page.getByLabel('Correo *').fill(correo);
    await page.getByLabel('Contraseña *').fill(clave);
    await page.getByRole('button', { name: 'Crear docente' }).click();
    await expect(page.getByText('La cuenta de docente se creó correctamente.')).toBeVisible();

    await page.getByRole('button', { name: 'Salir' }).click();
    await expect(page).toHaveURL(/\/ingreso$/);

    await page.getByLabel('Correo').fill(correo);
    await page.getByLabel('Contraseña').fill(clave);
    await page.getByRole('button', { name: 'Ingresar' }).click();
    await expect(page).toHaveURL(/\/asistencia$/);

    await page.goto('/alumnos');
    await expect(page.getByRole('heading', { name: 'Acceso no permitido' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Volver a mi inicio' })).toHaveAttribute('href', '/asistencia');
  });
});
