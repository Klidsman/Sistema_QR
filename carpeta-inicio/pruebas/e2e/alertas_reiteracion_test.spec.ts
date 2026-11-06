import { expect, test } from '@playwright/test';

import { configurarTurnoManana, crearAlumnoConQR, fechaEnLima, horaEnLima, ingresarComoAdministrador } from './ayudas';

test.describe('Historia 2: reiteración por tardanzas', () => {
  test('una tardanza con umbral uno genera la reiteración para el docente', async ({ page }) => {
    await ingresarComoAdministrador(page);

    await page.goto('/configuracion/umbrales');
    await page.getByLabel('Tipo *').selectOption({ label: 'Tardanzas' });
    await page.getByLabel('Valor *').fill('1');
    await page.getByLabel('Vigencia desde *').fill(fechaEnLima());
    await page.getByRole('button', { name: 'Guardar umbral' }).click();
    await expect(page.getByText('El umbral de alerta se guardó correctamente.')).toBeVisible();

    await configurarTurnoManana(page, horaEnLima(-30), horaEnLima(15), horaEnLima(45));

    const alumno = await crearAlumnoConQR(page);
    await page.goto('/porteria');
    await page.getByLabel('Código QR *').fill(alumno.token);
    await page.getByRole('button', { name: 'Registrar ingreso' }).click();
    await expect(page.getByText('Tardanza', { exact: true })).toBeVisible();

    await page.goto('/alertas');
    await expect(page.getByText(alumno.nombre, { exact: false }).first()).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Reiteración de tardanzas', exact: true }).first()).toBeVisible();
  });
});
