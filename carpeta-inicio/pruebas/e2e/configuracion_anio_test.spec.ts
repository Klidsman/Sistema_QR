import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Historia 1: año escolar, grados, secciones y turnos', () => {
  test('muestra el año activo y permite crear un grado', async ({ page }) => {
    await ingresarComoAdministrador(page);

    await page.goto('/configuracion/anio');
    await expect(page.getByRole('heading', { name: 'Año escolar' })).toBeVisible();
    await expect(page.getByText('Año activo')).toBeVisible();

    const sufijo = Date.now().toString().slice(-5);
    const grado = 'G' + sufijo;

    await page.goto('/configuracion/grados');
    await page.getByLabel('Nombre del grado *').fill(grado);
    await page.getByRole('button', { name: 'Agregar grado' }).click();

    await expect(page.getByText('El grado se creó correctamente.')).toBeVisible();
    await expect(page.getByText(grado, { exact: true })).toBeVisible();

    await page.getByLabel('Nombre del grado *').fill(grado);
    await page.getByRole('button', { name: 'Agregar grado' }).click();
    await expect(page.getByText('Ya existe un grado con ese nombre en el año.')).toBeVisible();
  });

  test('crea una sección y cambia su turno con vigencia', async ({ page }) => {
    await ingresarComoAdministrador(page);

    const sufijo = Date.now().toString().slice(-4);
    const grado = 'S' + sufijo;
    const seccion = 'Z' + sufijo;
    const hoy = new Date().toISOString().slice(0, 10);

    await page.goto('/configuracion/grados');
    await page.getByLabel('Nombre del grado *').fill(grado);
    await page.getByRole('button', { name: 'Agregar grado' }).click();
    await expect(page.getByText('El grado se creó correctamente.')).toBeVisible();

    await page.goto('/configuracion/secciones');
    await page.getByLabel('Grado *').selectOption({ label: grado });
    await page.getByLabel('Nombre de la sección *').fill(seccion);
    await page.getByLabel('Turno *').selectOption({ label: 'Mañana' });
    await page.getByLabel('Vigencia desde *').fill(hoy);
    await page.getByRole('button', { name: 'Agregar sección' }).click();

    await expect(page.getByText('La sección se creó correctamente.')).toBeVisible();
    const fila = page.locator('tr', { hasText: seccion });
    await expect(fila).toBeVisible();
    await expect(fila.locator('td').nth(2)).toHaveText('Mañana');

    await fila.getByLabel('Turno').selectOption({ label: 'Tarde' });
    await fila.getByLabel('Desde').fill(hoy);
    await fila.getByRole('button', { name: 'Cambiar' }).click();

    await expect(page.getByText('El turno de la sección se actualizó correctamente.')).toBeVisible();
    await expect(page.locator('tr', { hasText: seccion }).locator('td').nth(2)).toHaveText('Tarde');
  });
});
