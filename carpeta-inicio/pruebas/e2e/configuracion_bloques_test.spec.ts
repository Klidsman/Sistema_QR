import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function hoyISO(): string {
  return new Date().toISOString().slice(0, 10);
}

test.describe('Historia 6: bloques de hora de clase', () => {
  test('crea un bloque y rechaza uno solapado', async ({ page }) => {
    await ingresarComoAdministrador(page);

    await page.goto('/configuracion/turnos');
    const horario = page.locator('form[action="/configuracion/turnos/manana"]');
    await horario.getByLabel('Hora de ingreso (Mañana) *').fill('07:30');
    await horario.getByLabel('Tolerancia en minutos (Mañana) *').fill('10');
    await horario.getByLabel('Hora límite de tardanza (Mañana) *').fill('08:00');
    await horario.getByLabel('Hora de cierre (Mañana) *').fill('08:30');
    await horario.getByLabel('Vigencia desde (Mañana) *').fill(hoyISO());
    await horario.getByRole('button', { name: 'Guardar horario' }).click();
    await expect(page.getByText('El horario del turno se guardó correctamente.')).toBeVisible();

    await page.goto('/configuracion/bloques?turno=manana');
    await page.getByLabel('Número de bloque *').fill('1');
    await page.getByLabel('Hora de inicio *').fill('07:45');
    await page.getByLabel('Hora de fin *').fill('08:30');
    await page.getByLabel('Vigencia desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Agregar bloque' }).click();

    await expect(page.getByText('El bloque de clase se creó correctamente.')).toBeVisible();
    await expect(page.getByRole('cell', { name: '07:45', exact: true })).toBeVisible();

    await page.getByLabel('Número de bloque *').fill('2');
    await page.getByLabel('Hora de inicio *').fill('08:00');
    await page.getByLabel('Hora de fin *').fill('08:30');
    await page.getByLabel('Vigencia desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Agregar bloque' }).click();

    await expect(page.getByText('El bloque se superpone con otro bloque del mismo turno.')).toBeVisible();
  });
});
