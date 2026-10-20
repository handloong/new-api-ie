/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Switch } from '@/components/ui/switch'
import { getAdminPlans } from '@/features/subscriptions/api'
import type { PlanRecord } from '@/features/subscriptions/types'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const ldapSchema = z.object({
  LdapEnabled: z.boolean(),
  LdapServer: z.string().optional(),
  LdapPort: z.string().optional(),
  LdapUseTls: z.boolean(),
  LdapBindDn: z.string().optional(),
  LdapBindPassword: z.string().optional(),
  LdapBaseDn: z.string().optional(),
  LdapUserFilter: z.string().optional(),
  LdapAttrDisplayName: z.string().optional(),
  LdapAutoRegister: z.boolean(),
  LdapAutoRegisterPlanId: z.string().optional(),
})

type LdapFormValues = z.infer<typeof ldapSchema>

type LdapSectionProps = {
  defaultValues: LdapFormValues
}

export function LdapSection({ defaultValues }: LdapSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [plans, setPlans] = useState<PlanRecord[]>([])

  const form = useForm<LdapFormValues>({
    resolver: zodResolver(ldapSchema),
    defaultValues,
  })

  useEffect(() => {
    form.reset(defaultValues)
  }, [defaultValues, form])

  useEffect(() => {
    let cancelled = false
    getAdminPlans()
      .then((res) => {
        if (!cancelled && res.success && res.data) {
          setPlans(res.data)
        }
      })
      .catch(() => {
        // The plan list is optional; a failure only narrows the dropdown.
      })
    return () => {
      cancelled = true
    }
  }, [])

  const onSubmit = async (data: LdapFormValues) => {
    const updates = Object.entries(data).filter(
      ([key, value]) => value !== defaultValues[key as keyof LdapFormValues]
    )

    for (const [key, value] of updates) {
      await updateOption.mutateAsync({ key, value: value ?? '' })
    }
  }

  return (
    <SettingsSection title={t('LDAP Directory')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />

          <FormField
            control={form.control}
            name='LdapEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable LDAP')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Allow directory accounts to exchange credentials for an API token. No console session is created.'
                    )}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapServer'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Server')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='ldap://10.0.0.10'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Directory host or URL. The ldap:// and ldaps:// prefixes are accepted.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapPort'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Port')}</FormLabel>
                <FormControl>
                  <Input
                    inputMode='numeric'
                    placeholder='389'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Defaults to 389, or 636 when LDAPS is enabled.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapUseTls'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Use LDAPS')}</FormLabel>
                  <FormDescription>
                    {t('Connect to the directory over TLS.')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapBindDn'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Bind DN')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='CN=svc-ai,OU=Service,DC=corp,DC=local'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Service account used to search the directory. Leave empty for an anonymous bind.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapBindPassword'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Bind Password')}</FormLabel>
                <FormControl>
                  <Input
                    type='password'
                    autoComplete='new-password'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Password of the service account.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapBaseDn'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Base DN')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='OU=Users,DC=corp,DC=local'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Subtree searched for user entries.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapUserFilter'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('User Filter')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='(&(objectClass=user)(sAMAccountName={0}))'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('{0} is replaced with the escaped username.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapAttrDisplayName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Display Name Attribute')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='displayName'
                    {...field}
                    value={field.value ?? ''}
                  />
                </FormControl>
                <FormDescription>
                  {t('Attribute mapped to the display name, such as cn, en or displayName.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapAutoRegister'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Auto Register')}</FormLabel>
                  <FormDescription>
                    {t('Create a local account on the first successful directory login. Accounts are created without a password and cannot sign in to the console.')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='LdapAutoRegisterPlanId'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Subscription Plan')}</FormLabel>
                <FormControl>
                  <NativeSelect
                    {...field}
                    value={field.value ?? ''}
                    className='w-full'
                  >
                    <NativeSelectOption value=''>
                      {t('None - grant the default new-user quota')}
                    </NativeSelectOption>
                    {plans.map((record) => (
                      <NativeSelectOption
                        key={record.plan.id}
                        value={String(record.plan.id)}
                      >
                        {record.plan.title}
                        {record.plan.enabled ? '' : ` (${t('disabled')})`}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </FormControl>
                <FormDescription>
                  {t('Plan bound to newly registered accounts. When a plan is selected the default new-user quota is not granted.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
