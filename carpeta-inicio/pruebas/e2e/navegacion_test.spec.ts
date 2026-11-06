import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

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

    await navegacion.getByRole('link', { name: 'Docentes', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Docentes' })).toBeVisible();

    await navegacion.getByRole('link', { name: 'Alumnos', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Alumnos' })).toBeVisible();
  });
});
