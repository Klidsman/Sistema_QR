import { expect, test } from '@playwright/test';

import { configurarTurnoManana, crearAlumnoConQR, horaEnLima, ingresarComoAdministrador } from './ayudas';

test.describe('Historia 1: ingreso por QR en portería', () => {
  test('registra una tardanza y muestra la alerta al portero', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await configurarTurnoManana(page, horaEnLima(-30), horaEnLima(15), horaEnLima(45));

    const alumno = await crearAlumnoConQR(page);
    expect(alumno.token.length).toBeGreaterThan(0);

    await page.goto('/porteria');
    await page.getByLabel('Código QR *').fill(alumno.token);
    await page.getByRole('button', { name: 'Registrar ingreso' }).click();

    await expect(page.getByText('Tardanza', { exact: true })).toBeVisible();
    await expect(page.getByText(alumno.nombre, { exact: false })).toBeVisible();

    await page.goto('/alertas');
    await expect(page.getByText(alumno.nombre, { exact: false }).first()).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Tardanza', exact: true }).first()).toBeVisible();
  });

  test('rechaza un código inexistente', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/porteria');
    await page.getByLabel('Código QR *').fill('token-que-no-existe');
    await page.getByRole('button', { name: 'Registrar ingreso' }).click();

    await expect(page.getByText('El código QR no es válido.')).toBeVisible();
  });
});
