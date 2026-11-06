import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

const tamanos = [
  { nombre: 'celular', width: 390, height: 844 },
  { nombre: 'tablet', width: 820, height: 1180 },
  { nombre: 'computadora', width: 1366, height: 768 },
];

async function sinDesbordamientoHorizontal(page: import('@playwright/test').Page): Promise<boolean> {
  return page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1);
}

test.describe('Vista responsiva', () => {
  for (const tamano of tamanos) {
    test(`el listado y el formulario se usan en ${tamano.nombre}`, async ({ page }) => {
      await page.setViewportSize({ width: tamano.width, height: tamano.height });
      await ingresarComoAdministrador(page);

      await page.goto('/alumnos');
      await expect(page.getByRole('heading', { name: 'Alumnos' })).toBeVisible();
      await expect(page.getByLabel('Buscar')).toBeVisible();
      await expect(page.getByRole('button', { name: 'Filtrar' })).toBeVisible();
      expect(await sinDesbordamientoHorizontal(page)).toBe(true);

      await page.goto('/alumnos/nuevo');
      await expect(page.getByRole('heading', { name: 'Nuevo alumno' })).toBeVisible();
      await expect(page.getByLabel('DNI o código *')).toBeVisible();
      await expect(page.getByLabel('Grado y sección *')).toBeVisible();
      await expect(page.getByRole('button', { name: 'Guardar' })).toBeVisible();
      expect(await sinDesbordamientoHorizontal(page)).toBe(true);
    });
  }
});
